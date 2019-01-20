FROM nixos/nix:2.1.3

RUN apk add --no-cache git bash openssh && \
    . "$HOME/.nix-profile/etc/profile.d/nix.sh" && \
    nix-env -if https://github.com/cachix/cachix/tarball/master \
        --substituters https://cachix.cachix.org \
        --trusted-public-keys cachix.cachix.org-1:eWNHQldwUO7G2VkjpnjDbWwy4KQ/HNxht7H4SSoMckM= && \
    cachix use dapp && \
    git clone --recursive https://github.com/dapphub/dapptools $HOME/.dapp/dapptools && \
    nix-env -f $HOME/.dapp/dapptools -iA dapp seth solc hevm ethsign && \
    rm -rf /var/cache/apk/*
