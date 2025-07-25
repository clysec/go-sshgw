# Go SSH Gateway
This is a simple SSH gateway setup using Go and HAProxy. It allows you to connect to multiple SSH servers through a single TLS entrypoint.

## Installation
### Binaries
You can fetch the pre-built binaries for your platform from the [releases page](https://github.com/clysec/go-sshgw/releases).

### Building from Source
To build the binary from source, you need to have Go 1.24.3 installed. Clone the repository and run the following commands:

```bash
git clone https://github.com/clysec/go-sshgw.git
cd go-sshgw
go build -o sshgw
```

### Debian Package
```
# Ensure prerequisites are installed
sudo apt-get update 
sudo apt-get -y install gnupg2 apt-transport-https curl

# Download and import the public key for verifying packages
sudo curl https://pkg.cloudyne.io/debian/repository.key -o /etc/apt/keyrings/cydeb.asc

# Add the package repository
echo "deb [signed-by=/etc/apt/keyrings/cydeb.asc] https://pkg.cloudyne.io/debian all main" | sudo tee -a /etc/apt/sources.list.d/cydeb.list

# Check that the repository works
sudo apt-get update

# Install the sshgw package
sudo apt-get install sshgw
```

### RPM Package
```bash
# on RedHat based distributions
dnf config-manager --add-repo https://git.cloudyne.io/api/packages/linux/rpm.repo

# on SUSE based distributions
zypper addrepo https://git.cloudyne.io/api/packages/linux/rpm.repo

# Install
dnf install sshgw --nogpgcheck
```

### Alpine
```bash
# get and save the key
cd /etc/apk/keys && curl -JO https://git.cloudyne.io/api/packages/linux/alpine/key 

# Add the following line to /etc/apk/repositories
echo 'https://git.cloudyne.io/api/packages/linux/alpine/all/repository' >> /etc/apk/repositories

# Update
apk update

# Install
apk add sshgw
```

## SSH Configuration
### Linux/Darwin
First, download the `sshgw` binary and place it in your desired directory, for example `/usr/local/bin/sshgw`.

To configure your SSH client to connect through the gateway, you can add the following lines to your `~/.ssh/config` file:

```ssh
Host my.ssh.server.com
  ProxyCommand /path/to/sshgw sshproxy.example.com:22 my.ssh.server.com
  User myuser
  IdentityFile ~/.ssh/id_rsa
```

You can then connect to the server using:

```bash
ssh my.ssh.server.com
```

### Windows
For Windows users, you can use the `sshgw.exe` binary in a similar way. Make sure to place it in a directory included in your PATH, or specify the full path in your SSH configuration.
To configure your SSH client, you can create or edit the `C:\Users\<YourUsername>\.ssh\config` file and add the following lines:

```ssh
Host my.ssh.server.com
  ProxyCommand C:\\path\\to\\sshgw.exe sshproxy.example.com:22 my.ssh.server.com
  User myuser
  IdentityFile C:\\Users\\<YourUsername>\\.ssh\\id_rsa
```

You can then connect to the server using:

```bash
ssh my.ssh.server.com
```

## HAProxy Configuration
### Specific Servers
This setup will allow you to connect to the servers "serverA" and "serverB" through the gateway. The configuration is done in the `haproxy.cfg` file.

```config
frontend ssh_gateway
    bind *:443 ssl crt /etc/haproxy/certs/sshgw.pem
    mode tcp
    option tcplog

    tcp-request content set-var(sess.dstName) ssl_fc_sni
    log-format "%ci:%cp [%t] %ft %b/%s %Tw/%Tc/%Tt %B %ts %ac/%fc/%bc/%sc/%rc %sq/%bq dstName:%[var(sess.dstName)]"

    use_backend serverA if { var(sess.dstName) -m str serverA.example.com }
    use_backend serverB if { var(sess.dstName) -m str serverB.example.com }

backend serverA
    mode tcp
    server serverA serverA.example.com:22

backend serverB
    mode tcp
    server serverB serverB.example.com:22
```

### Generic Gateway (no restrictions)
If you want to create a generic SSH gateway that can handle any SSH server without specific restrictions, you can use the following configuration:
```
resolvers internal
   accepted_payload_size 16384
   nameserver dns1      1.1.1.1:53
   resolve_retries      4
   timeout resolve      1s
   timeout retry        3s
   hold other           30s
   hold refused         30s
   hold nx              30s
   hold timeout         30s
   hold valid           10s
   hold obsolete        30s

defaults
    log     global
    mode    tcp
    option  tcplog
    option  dontlognull
    timeout connect 5000
    timeout client  50000
    timeout server  50000

frontend ssh_gateway
   bind *:443 ssl crt /etc/haproxy/certs/sshgw.pem
   mode tcp
   
   tcp-request content do-resolve(sess.dstIP,internal,ipv4) ssl_fc_sni
   tcp-request content set-var(sess.dstName) ssl_fc_sni

   log-format "%ci:%cp [%t] %ft %b/%s %Tw/%Tc/%Tt %B %ts %ac/%fc/%bc/%sc/%rc %sq/%bq dstName:%[var(sess.dstName)] dstIP:%[var(sess.dstIP)] "

   default_backend all-ssh
   
backend all-ssh
  mode tcp
  
  server all-ssh %[var(sess.dstIP)]:22 resolvers internal check resolvers internal

frontend ssh_gateway
   bind *:443 ssl crt /etc/haproxy/certs/sshgw.pem
   mode tcp
   
   tcp-request inspect-delay 2s
   tcp-request content do-resolve(sess.dstIP,internal,ipv4) ssl_fc_sni
   tcp-request content set-var(sess.dstName) ssl_fc_sni

   log-format "%ci:%cp [%t] %ft %b/%s %Tw/%Tc/%Tt %B %ts %ac/%fc/%bc/%sc/%rc %sq/%bq dstName:%[var(sess.dstName)] dstIP:%[var(sess.dstIP)] "

   default_backend all-ssh
   
backend all-ssh
  mode tcp

  tcp-request content set-dst var(sess.dstIP)

  server all-ssh 0.0.0.0:22
```