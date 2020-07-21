/**************************************************************************
 * client.c                                                               *
 *                                                                        *
 *************************************************************************/

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <net/if.h>
#include <linux/if_tun.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/ioctl.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <arpa/inet.h>
#include <sys/select.h>
#include <sys/time.h>
#include <errno.h>
#include <stdarg.h>
#include <netinet/tcp.h>

#define PORT 55555
#define MSS_TCP 500

 struct ip_option_header
 {
   uint8_t type;
   uint8_t length;
   uint8_t pointer;
   uint8_t padding;
   uint8_t route_data[36];
 } option_data;

int debug;
char *progname;

/**************************************************************************
 * cread: read routine that checks for errors and exits if an error is    *
 *        returned.                                                       *
 **************************************************************************/
int cread(int fd, char *buf, int n){

  int nread;

  if((nread=read(fd, buf, n)) < 0){
    perror("Reading data");
    exit(1);
  }
  return nread;
}

/**************************************************************************
 * cwrite: write routine that checks for errors and exits if an error is  *
 *         returned.                                                      *
 **************************************************************************/
int cwrite(int fd, char *buf, int n){

  int nwrite;

  if((nwrite=write(fd, buf, n)) < 0){
    perror("Writing data");
    exit(1);
  }
  return nwrite;
}

/**************************************************************************
 * read_n: ensures we read exactly n bytes, and puts them into "buf".     *
 *         (unless EOF, of course)                                        *
 **************************************************************************/
int read_n(int fd, char *buf, int n) {

  int nread, left = n;

  while(left > 0) {
    if ((nread = cread(fd, buf, left)) == 0){
      return 0 ;
    }else {
      left -= nread;
      buf += nread;
    }
  }
  return n;
}

/**************************************************************************
 * do_debug: prints debugging stuff (doh!)                                *
 **************************************************************************/
void do_debug(char *msg, ...){

  va_list argp;

  if(debug) {
	va_start(argp, msg);
	vfprintf(stderr, msg, argp);
	va_end(argp);
  }
}

/**************************************************************************
 * my_err: prints custom error messages on stderr.                        *
 **************************************************************************/
void my_err(char *msg, ...) {

  va_list argp;

  va_start(argp, msg);
  vfprintf(stderr, msg, argp);
  va_end(argp);
}

/**************************************************************************
 * usage: prints usage and exits.                                         *
 **************************************************************************/
void usage(void) {
  fprintf(stderr, "Usage:\n");
  fprintf(stderr, "%s [-c <serverIP>] [-p <port>] [-d]\n", progname);
  fprintf(stderr, "%s -h\n", progname);
  fprintf(stderr, "\n");
  fprintf(stderr, "-c <serverIP>: server address (mandatory)\n");
  fprintf(stderr, "-p <port>: port to connect to, default 55555\n");
  fprintf(stderr, "-d: outputs debug information while running\n");
  fprintf(stderr, "-h: prints this help text\n");
  exit(1);
}

int main(int argc, char *argv[]) {

  int option;
  uint16_t nwrite, plength, nread = 0;
  char buffer[MSS_TCP];
  struct sockaddr_in local, remote;
  char remote_ip[16] = "";            /* dotted quad IP string */
  unsigned short int port = PORT;
  int sock_fd, net_fd, optval = 1;
  socklen_t remotelen;

  progname = argv[0];

  /* Check command line options */
  while((option = getopt(argc, argv, "c:p:hd")) > 0) {
    switch(option) {
      case 'd':
        debug = 1;
        break;
      case 'h':
        usage();
        break;
      case 'c':
        strncpy(remote_ip,optarg,15);
        break;
      case 'p':
        port = atoi(optarg);
        break;
      default:
        my_err("Unknown option %c\n", option);
        usage();
    }
  }

  argv += optind;
  argc -= optind;

  if(argc > 0) {
    my_err("Too many options!\n");
    usage();
  }

  if(*remote_ip == '\0') {
    my_err("Must specify server address!\n");
    usage();
  }

  if ( (sock_fd = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
    perror("socket()");
    exit(1);
  }

  bzero((char *)&option_data, sizeof(option_data));
  option_data.type          = 68;
  option_data.length        = 40;
  option_data.pointer       = 5;

  ret = setsockopt(fd_socket, IPPROTO_IP, IP_OPTIONS, (char *)&option_data, sizeof(option_data));
  if(ret != 0) { printf("tcpclient.c - Setup of socket options failed: %s.\n", strerror(errno)); exit(0); }
  else { printf("tcpclient.c - Setup of socket options successfull.\n"); }


  /* set the maximum segment size */
  optval = MSS_TCP;
  if(setsockopt(sock_fd, IPPROTO_TCP, TCP_MAXSEG, (char *)&optval, sizeof(optval)) < 0) {
    perror("setsockopt()");
    exit(1);
  }

  /* assign the destination address */
  memset(&remote, 0, sizeof(remote));
  remote.sin_family = AF_INET;
  remote.sin_addr.s_addr = inet_addr(remote_ip);
  remote.sin_port = htons(port);

  /* connection request */
  if (connect(sock_fd, (struct sockaddr*) &remote, sizeof(remote)) < 0) {
    perror("connect()");
    exit(1);
  }

  net_fd = sock_fd;
  do_debug("CLIENT: Connected to server %s\n", inet_ntoa(remote.sin_addr));

  int counter = 0;
  while(counter < 100) {

    counter++;
    //sleep(2);

    int index = 1;
    for(int i = 0; i < MSS_TCP; i++) {
      if(i < MSS_TCP - 9)
      {
        buffer[i] = 0;
      }
      else
      {
        buffer[i] = index;
        index++;
      }
    }

    /* write length + packet */
    //plength = htons(nread);
    //nwrite = cwrite(net_fd, (char *)&plength, sizeof(plength));
    nwrite = cwrite(net_fd, buffer, sizeof(buffer));

  }

  return(0);
}
