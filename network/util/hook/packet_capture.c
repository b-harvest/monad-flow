#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/pkt_cls.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define MAX_PKT_SIZE 65535

struct pkt_sample {
    __u32 len;
    char data[MAX_PKT_SIZE];
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} events SEC(".maps");

char LICENSE[] SEC("license") = "GPL";

static __always_inline int parse_packet(struct __sk_buff *skb) {
    void *data_end = (void *)(long)skb->data_end;
    void *data = (void *)(long)skb->data;

    // 1. L2 (Ethernet) header parse
    struct ethhdr *eth = data;
    if ((void *)eth + sizeof(*eth) > data_end) {
        return TC_ACT_OK;
    }

    // 2. L3 (IP) header parse
    if (eth->h_proto != bpf_htons(ETH_P_IP)) {
        return TC_ACT_OK;
    }
    struct iphdr *iph = data + sizeof(*eth);
    if ((void *)iph + sizeof(*iph) > data_end) {
        return TC_ACT_OK;
    }

    // 3. L4 (TCP/UDP) header parse
    __u16 sport, dport;
    if (iph->protocol == IPPROTO_TCP) {
        struct tcphdr *tcph = (void *)iph + sizeof(*iph);
        if ((void *)tcph + sizeof(*tcph) > data_end) {
            return TC_ACT_OK;
        }
        sport = bpf_ntohs(tcph->source);
        dport = bpf_ntohs(tcph->dest);
    } else if (iph->protocol == IPPROTO_UDP) {
        struct udphdr *udph = (void *)iph + sizeof(*iph);
        if ((void *)udph + sizeof(*udph) > data_end) {
            return TC_ACT_OK;
        }
        sport = bpf_ntohs(udph->source);
        dport = bpf_ntohs(udph->dest);
    } else {
        return TC_ACT_OK;
    }

    // port filtering
    if (sport == 8000 || dport == 8000) {
        bpf_printk("[MONAD DEBUG] Proto: %u, SRC IP: %pI4, DST IP: %pI4, SRC Port: %u, DST Port: %u\n",
                   iph->protocol, &iph->saddr, &iph->daddr, sport, dport);

        // 1. 상수 크기(sizeof(struct pkt_sample))를 예약합니다.
        struct pkt_sample *sample = bpf_ringbuf_reserve(&events, sizeof(struct pkt_sample), 0);
        if (!sample) {
            bpf_printk("[MONAD DEBUG] Ringbuf reserve failed\n");
            return TC_ACT_OK;
        }

        // 2. 실제 패킷 길이를 구조체에 저장합니다.
        __u32 pkt_len = skb->len;
        sample->len = pkt_len;

        // 3. 패킷 데이터를 구조체의 data 필드로 복사합니다.
        __u32 copy_len = (pkt_len < MAX_PKT_SIZE) ? pkt_len : MAX_PKT_SIZE;
        
        if (copy_len > 0) {
            if (bpf_skb_load_bytes(skb, 0, sample->data, copy_len) < 0) {
                bpf_printk("[MONAD DEBUG] skb_load_bytes failed\n");
                bpf_ringbuf_discard(sample, 0);
                return TC_ACT_OK;
            }
        }

        // application readable
        bpf_ringbuf_submit(sample, 0);
    }

    return TC_ACT_OK;
}

// TC Ingress recv hook
SEC("classifier")
int tc_ingress(struct __sk_buff *skb) {
    return parse_packet(skb);
}

// TC Egress send hook
SEC("classifier")
int tc_egress(struct __sk_buff *skb) {
    return parse_packet(skb);
}