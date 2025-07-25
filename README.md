# Go SSH Gateway
This is a simple SSH gateway setup using Go and HAProxy. It allows you to connect to multiple SSH servers through a single TLS entrypoint.

## SSH Configuration
### Linux
To configure your SSH client to connect through the gateway, you can add the following lines to your `~/.ssh/config` file:

```ssh
Host server-to-connect
  ProxyCommand /path/to/sshgw sshproxy.example.com:22 %h
```

You can then connect to the server using:

```bash
ssh server-to-connect
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