#!/bin/bash
set -e

# EXPECTS .sops.yaml TO SPECIFY CREATION RULES
echo "Verifying PGP key is imported..."
gpg --import test/key.asc
