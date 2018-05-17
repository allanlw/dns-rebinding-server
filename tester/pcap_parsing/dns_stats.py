from scapy.all import *
import numpy as np
import matplotlib.pyplot as plt
import sys
cap = rdpcap(sys.argv[1])
dnspackets = []
for packet in cap:
    if packet.lastlayer().name == 'DNS':
        dnspackets.append(packet)
def dns_timing():
    times=[]
    for packet in dnspackets:
        times.append(packet.time)
    times = sorted(times)
    times = [x - times[0] for x in times]

    n, bins, patches = plt.hist(times,50,density=True,facecolor='g',alpha=0.75)
    plt.show()


def dns_interarrival():
    names = defaultdict(list)
    for packet in dnspackets:
        qname = packet.lastlayer().qd.qname
        names[qname].append(packet)
    interarrival_times = []
    for name in names:
        packets = names[name]
        query_times = [packet.time for packet in packets if packet.lastlayer().an == None]
        diffs = []
        for cnt,time in enumerate(query_times):
            if cnt <= len(query_times) - 2:
                diff = (query_times[cnt] - query_times[cnt+1])
                diffs.append(abs(diff))
        #diffs = [x for x in diffs if x > 0.1]
        interarrival_times.append((name,diffs))
    n,bins,patches = plt.hist([x[1] for x in interarrival_times],50,density=True)
    plt.show()
dns_interarrival()
