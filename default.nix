{}:

let
    pkgs = import <nixpkgs> {};
    envPkgs = with pkgs;
    [
     direnv
     coreutils
     bash
     #stdenv.shell
     # zsh
     #curl
     # git
    ];
    path = "PATH=/usr/bin:/bin";
in
    pkgs.dockerTools.buildLayeredImage {
        maxLayers = 104;
        name = "cmgolang";
        tag = "latest";
        contents = envPkgs;
        extraCommands = ''
            mkdir ./cm
            ln -s "${pkgs.bash}/bin/bash" ./bash
            ln -s "${pkgs.bash}/bin/bash" ./cm/bash
        '';
        config = ({
            Cmd = [ "/bin/echo" "hello" ];
            Env = [ path ];
            }
        );
    }
