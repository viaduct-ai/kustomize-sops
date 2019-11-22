set -e

KUSTOMIZE="kustomize"

if [ -x "$(command -v $KUSTOMIZE)" ]; then
  KUSTOMIZE_EXEC=$(command -v $KUSTOMIZE)
  # Remove existing kustomize executable
  echo "Removing existing $KUSTOMIZE executable..."
  echo "rm $KUSTOMIZE_EXEC"
  rm $KUSTOMIZE_EXEC
fi

echo "Installing $KUSTOMIZE..."
go install sigs.k8s.io/kustomize/v3/cmd/kustomize

echo "Successfully reinstalled $KUSTOMIZE!"
kustomize version
