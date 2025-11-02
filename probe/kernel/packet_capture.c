#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/in.h>

#define TARGET_PORT 8000

SEC("socket")
int bpf_filter(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    // 1. 이더넷 헤더 (IP 패킷인지 확인)
    struct ethhdr *eth = data;
    if ((void *)eth + sizeof(*eth) > data_end) {
        return 0; // Drop
    }
    if (eth->h_proto != bpf_htons(ETH_P_IP)) {
        return 0; // Not IP
    }

    // 2. IP 헤더 (TCP/UDP인지 확인)
    struct iphdr *ip = data + sizeof(*eth);
    if ((void *)ip + sizeof(*ip) > data_end) {
        return 0;
    }
    if (ip->protocol != IPPROTO_TCP && ip->protocol != IPPROTO_UDP) {
        return 0; // Not TCP or UDP
    }

    // 3. TCP/UDP 헤더 (포트 번호 확인)
    void *transport_header = (void *)ip + (ip->ihl * 4);

    // 8000번 포트를 네트워크 바이트 순서로 변환
    __u16 target_port_net = bpf_htons(TARGET_PORT);
    __u16 src_port = 0;
    __u16 dst_port = 0;

    if (ip->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = transport_header;
        if ((void *)tcp + sizeof(*tcp) > data_end) {
            return 0;
        }
        src_port = tcp->source;
        dst_port = tcp->dest;
    } else { // UDP
        struct udphdr *udp = transport_header;
        if ((void *)udp + sizeof(*udp) > data_end) {
            return 0;
        }
        src_port = udp->source;
        dst_port = udp->dest;
    }

    // 4. 필터링 로직
    if (src_port == target_port_net || dst_port == target_port_net) {
        // 포트 8000번이 맞으면 패킷을 통과시킴
        return -1;
    }

    return 0; // 그 외 모든 패킷은 차단
}

char _license[] SEC("license") = "GPL";