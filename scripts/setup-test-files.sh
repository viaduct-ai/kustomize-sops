#!/bin/bash
set -e


# EXPECTS .sops.yaml TO SPECIFY CREATION RULES

TEST_DIR="test/"
SECRET="secret"
YAML_SUFFIX=".yaml"
ENC_SUFFIX=".enc"

echo "Verifying PGP key is imported..."
gpg --import ${TEST_DIR}/key.asc

echo "Generating test files..."
sops -e ${TEST_DIR}${SECRET}.yaml > ${TEST_DIR}${SECRET}${ENC_SUFFIX}${YAML_SUFFIX}
sops -e ${TEST_DIR}${SECRET}-A.yaml > ${TEST_DIR}${SECRET}-A${ENC_SUFFIX}${YAML_SUFFIX}
sops -e ${TEST_DIR}${SECRET}-B.yaml > ${TEST_DIR}${SECRET}-B${ENC_SUFFIX}${YAML_SUFFIX}
sops -e ${TEST_DIR}${SECRET}-C.yaml > ${TEST_DIR}${SECRET}-C${ENC_SUFFIX}${YAML_SUFFIX}
