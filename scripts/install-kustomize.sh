set -e

KUSTOMIZE="kustomize"

function install_kustomize() {
  echo "Installing $KUSTOMIZE..."
  go install sigs.k8s.io/kustomize/v3/cmd/kustomize

  echo "Successfully reinstalled $KUSTOMIZE!"
  kustomize version
}

if [ -x "$(command -v $KUSTOMIZE)" ]; then
  KUSTOMIZE_EXEC=$(command -v $KUSTOMIZE)

  echo "WARNING: Found an existing installation of $KUSTOMIZE at $KUSTOMIZE_EXEC"
  read -p "Please confirm you want to reinstall $KUSTOMIZE (y/n): " -n 1 -r
  echo 
  if [[ $REPLY =~ ^[Yy]$ ]]
  then
      # do dangerous stuff
    echo "Removing existing $KUSTOMIZE executable..."
    # Remove existing kustomize executable
    echo "rm $KUSTOMIZE_EXEC"
    rm $KUSTOMIZE_EXEC

    # Install
    install_kustomize
  fi
else
    # Install
    install_kustomize
fi

