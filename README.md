# esshd
Ephemeral SSHD for containers

# NOTE:
# !!! NOT FOR SECURE LOGINS !!!!

# Usage

Set entrypoint to `esshd` executable location.
Argument #1 set to the binary to execute upon SSH login.
Argument #2 set to the host:port.

Example:

    /esshd /bin/bash 127.0.0.1:2222

# Banner

If /esshd.txt is found inside the container image then it will be used as the SSH login banner.
