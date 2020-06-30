#include <stdio.h>
#include <stdlib.h>
#include <netinet/tcp.h>

struct tcp_info tcp_info;
int tcp_info_length;

int main ( int argc, char **argv ) {

    if (argc != 2) {
        printf("Expect exactly one argument!\n");
        return 1;
    }

    int fd = atoi(argv[1]);

    tcp_info_length = sizeof(tcp_info);
    if ( getsockopt( fd, SOL_TCP, TCP_INFO, (void *)&tcp_info, &tcp_info_length ) == 0 ) {
    	printf("%u %u %u %u %u %u %u %u %u %u %u %u\n",
    			tcp_info.tcpi_last_data_sent,
    			tcp_info.tcpi_last_data_recv,
    			tcp_info.tcpi_snd_cwnd,
    			tcp_info.tcpi_snd_ssthresh,
    			tcp_info.tcpi_rcv_ssthresh,
    			tcp_info.tcpi_rtt,
    			tcp_info.tcpi_rttvar,
    			tcp_info.tcpi_unacked,
    			tcp_info.tcpi_sacked,
    			tcp_info.tcpi_lost,
    			tcp_info.tcpi_retrans,
    			tcp_info.tcpi_fackets);
    };
}