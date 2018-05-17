Project to implement a DNS server and client program to facilitate DNS rebinding attacks.
Contact: dcooper@akamai.com, awirth@akamai.com

To build:

	docker build -t dnsrebind .

To run (with http server on localhost:8080 in the host):

	docker run -it --rm -v $(pwd)/conf:/etc/dnsrebind/ -p80:80 -p53:8053/udp -p53:8053/tcp -p10000:10000 dnsrebind

This will use the configuration file in `$(pwd)/conf/dnsrebind.yml`.

There's a format script. you need `sudo npm install -g js-beautify standard` to have the right tools.

## IPTables Voodoo

We want a secondary IP address to slurp all traffic on all ports. We need to add a secondary IP to the server. Follow this
Stack Exchange answer: https://serverfault.com/a/864615

Then, we need to write an iptables rule to redirect the traffic:

    sudo iptables -I PREROUTING -t nat -i eth0 -d 172.31.36.162 -p tcp --dport 1:65535 -j DNAT --to-destination 172.31.39.125:10000

Note, replace the first and second IP addresses with the appropriate internal IPs.

To inspect the rule, use:

    sudo iptables -L -t nat

# License

Copyright 2018 Akamai Technologies, Inc

Permission is hereby granted, free of charge, to any person obtaining a copy of this software 
and associated documentation files (the "Software"), to deal in the Software without 
restriction, including without limitation the rights to use, copy, modify, merge, publish, 
distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the 
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or 
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING 
BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND 
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, 
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, 
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
