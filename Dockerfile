FROM nixos/nix:2.3.6

COPY build/linux/up /bin/

RUN apk add --no-cache --update \
&& apk add curl

RUN nix-env -iA \
 nixpkgs.go_1_14 \
 nixpkgs.jq \
 nixpkgs.yq-go \
 nixpkgs.lua \
 nixpkgs.direnv \
 nixpkgs.coreutils \
 nixpkgs.zsh \
 nixpkgs.tzdata \
 nixpkgs.git \
 nixpkgs.vim \
 nixpkgs.fd \
 nixpkgs.which \
 nixpkgs.ripgrep \
 nixpkgs.gnugrep \
 nixpkgs.gawk \
 nixpkgs.findutils \
 nixpkgs.vifm \
 nixpkgs.fzf \
 nixpkgs.highlight \
 nixpkgs.universal-ctags \
 nixpkgs.readline \
 nixpkgs.tree \
 nixpkgs.oh-my-zsh \
 nixpkgs.wget \
 nixpkgs.zsh \
 nixpkgs.docker \
 nixpkgs.awscli \
 nixpkgs.fzf \
 nixpkgs.ncdu \
 nixpkgs.helm \
 nixpkgs.gnused \
 nixpkgs.pup \
 nixpkgs.modd
