#!/usr/bin/env bash
# NexaRail mainnet sentry — Oracle Cloud bootstrap.
# Run inside Oracle Cloud Shell (pre-authenticated OCI CLI).
# Provisions: VCN + subnet + IGW + security list with SSH + Tendermint p2p,
# then launches an Always Free Ampere A1.Flex VM (4 OCPU, 24 GB, 100 GB boot)
# pre-authorized with our SSH public key.

set -Eeuo pipefail

NAME="nexarail-sentry-1"
P2P_PORT=32656

# SSH public key — generated 2026-06-21 on Bradley's Mac (~/.ssh/nexarail-sentry-oracle).
SSH_PUB="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMjeULcMG2XrgPi1LerE4ufan0GfdmWTmw1camRCEIKN nexarail-sentry-oracle-2026-06-21"

# Tenancy OCID — Cloud Shell sets OCI_TENANCY automatically.
if [ -z "${OCI_TENANCY:-}" ]; then
  echo "ERROR: \$OCI_TENANCY is not set."
  echo
  echo "This script must run inside Oracle Cloud Shell (browser terminal in the"
  echo "OCI Console), not your local Mac terminal."
  echo
  echo "  1. Open https://cloud.oracle.com"
  echo "  2. Click the '>_' icon (top-right, near your avatar) to open Cloud Shell."
  echo "  3. Re-paste:"
  echo "     curl -fsSL https://raw.githubusercontent.com/Bookings-cpu/nexarail/main/scripts/mainnet/oracle-sentry-bootstrap.sh | bash"
  exit 2
fi
COMPARTMENT_ID="$OCI_TENANCY"
echo "tenancy/root compartment: $COMPARTMENT_ID"

REGION="$(oci iam region-subscription list \
  --query 'data[?"is-home-region"==`true`]."region-name" | [0]' --raw-output)"
echo "home region: $REGION"

# Pick the latest Ubuntu 22.04 ARM image.
IMAGE_ID="$(oci compute image list \
  --compartment-id "$COMPARTMENT_ID" \
  --operating-system "Canonical Ubuntu" \
  --operating-system-version "22.04" \
  --shape "VM.Standard.A1.Flex" \
  --sort-by TIMECREATED --sort-order DESC --limit 1 \
  --query 'data[0].id' --raw-output)"
echo "image: $IMAGE_ID"

# VCN — reuse existing or create.
VCN_ID="$(oci network vcn list --compartment-id "$COMPARTMENT_ID" \
  --display-name "nexarail-vcn" --query 'data[0].id' --raw-output 2>/dev/null || true)"
if [ -z "${VCN_ID:-}" ] || [ "$VCN_ID" = "null" ]; then
  echo "creating VCN..."
  VCN_ID="$(oci network vcn create --compartment-id "$COMPARTMENT_ID" \
    --cidr-block "10.0.0.0/16" --display-name "nexarail-vcn" --dns-label "nexarail" \
    --wait-for-state AVAILABLE --query 'data.id' --raw-output)"

  IGW_ID="$(oci network internet-gateway create --compartment-id "$COMPARTMENT_ID" \
    --vcn-id "$VCN_ID" --display-name "nexarail-igw" --is-enabled true \
    --wait-for-state AVAILABLE --query 'data.id' --raw-output)"

  DRT_ID="$(oci network vcn get --vcn-id "$VCN_ID" \
    --query 'data."default-route-table-id"' --raw-output)"
  oci network route-table update --rt-id "$DRT_ID" --force \
    --route-rules "[{\"destination\":\"0.0.0.0/0\",\"destinationType\":\"CIDR_BLOCK\",\"networkEntityId\":\"$IGW_ID\"}]" \
    --wait-for-state AVAILABLE >/dev/null

  DSL_ID="$(oci network vcn get --vcn-id "$VCN_ID" \
    --query 'data."default-security-list-id"' --raw-output)"
  oci network security-list update --security-list-id "$DSL_ID" --force \
    --ingress-security-rules "[
      {\"source\":\"0.0.0.0/0\",\"protocol\":\"6\",\"isStateless\":false,\"tcpOptions\":{\"destinationPortRange\":{\"min\":22,\"max\":22}}},
      {\"source\":\"0.0.0.0/0\",\"protocol\":\"6\",\"isStateless\":false,\"tcpOptions\":{\"destinationPortRange\":{\"min\":${P2P_PORT},\"max\":${P2P_PORT}}}}
    ]" --wait-for-state AVAILABLE >/dev/null
fi
echo "vcn: $VCN_ID"

# Subnet — reuse existing or create.
SUBNET_ID="$(oci network subnet list --compartment-id "$COMPARTMENT_ID" \
  --vcn-id "$VCN_ID" --display-name "nexarail-subnet" \
  --query 'data[0].id' --raw-output 2>/dev/null || true)"
if [ -z "${SUBNET_ID:-}" ] || [ "$SUBNET_ID" = "null" ]; then
  SUBNET_ID="$(oci network subnet create --compartment-id "$COMPARTMENT_ID" \
    --vcn-id "$VCN_ID" --cidr-block "10.0.0.0/24" \
    --display-name "nexarail-subnet" --dns-label "sentry" \
    --wait-for-state AVAILABLE --query 'data.id' --raw-output)"
fi
echo "subnet: $SUBNET_ID"

# Write SSH key to a temp file (oci CLI needs a path).
KEY_FILE="$(mktemp)"
printf '%s\n' "$SSH_PUB" > "$KEY_FILE"

# Launch — try each AD in order, capacity errors retry next AD.
ADS=( $(oci iam availability-domain list --compartment-id "$COMPARTMENT_ID" \
  --query 'data[].name' --raw-output | tr -d '[]," ' | tr '\n' ' ') )
INSTANCE_ID=""
for AD in "${ADS[@]}"; do
  [ -z "$AD" ] && continue
  echo "attempting launch in AD: $AD"
  if INSTANCE_ID="$(oci compute instance launch \
      --compartment-id "$COMPARTMENT_ID" \
      --availability-domain "$AD" \
      --shape "VM.Standard.A1.Flex" \
      --shape-config '{"ocpus":4,"memoryInGBs":24}' \
      --image-id "$IMAGE_ID" \
      --subnet-id "$SUBNET_ID" \
      --display-name "$NAME" \
      --assign-public-ip true \
      --boot-volume-size-in-gbs 100 \
      --ssh-authorized-keys-file "$KEY_FILE" \
      --wait-for-state RUNNING \
      --query 'data.id' --raw-output 2>&1)"; then
    echo "launched in $AD"
    break
  else
    echo "launch failed in $AD:"
    echo "$INSTANCE_ID" | head -10
    INSTANCE_ID=""
  fi
done

rm -f "$KEY_FILE"

if [ -z "$INSTANCE_ID" ]; then
  echo
  echo "ERROR: could not provision Ampere A1 in any AD."
  echo "Most likely cause: Always Free Ampere capacity exhausted in this region."
  echo "Either try again in a few hours, or switch to AMD Always Free (E2.1.Micro)."
  exit 1
fi

# Public IP
sleep 5
PUBLIC_IP="$(oci compute instance list-vnics --instance-id "$INSTANCE_ID" \
  --query 'data[0]."public-ip"' --raw-output)"

echo
echo "=========================================="
echo "  NEXARAIL SENTRY VM PROVISIONED"
echo "=========================================="
echo "  instance:  $INSTANCE_ID"
echo "  region:    $REGION"
echo "  public IP: $PUBLIC_IP"
echo "  ssh:       ssh -i ~/.ssh/nexarail-sentry-oracle ubuntu@$PUBLIC_IP"
echo "=========================================="
