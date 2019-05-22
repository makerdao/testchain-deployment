FROM nixos/nix:2.1.3

ENV XDG_CACHE_HOME=/nix/cache

RUN apk add --no-cache git bash openssh && \
    . "$HOME/.nix-profile/etc/profile.d/nix.sh" && \
    nix run -f https://github.com/cachix/cachix/tarball/master \
        --substituters https://cachix.cachix.org \
        --trusted-public-keys cachix.cachix.org-1:eWNHQldwUO7G2VkjpnjDbWwy4KQ/HNxht7H4SSoMckM= \
        -c sh -c "cachix use dapp && cachix use tdds" && \
    nix-env -f https://github.com/makerdao/nixpkgs-pin/tarball/master -iA pkgs.makerCommonScriptBins && \
    rm -rf /var/cache/apk/* && \
    nix-collect-garbage -d && \
    nix-store --optimise && \
    nix-env -q

VOLUME /nix
