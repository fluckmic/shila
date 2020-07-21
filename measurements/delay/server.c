/**************************************************************************
 * server.c                                                               *
 *                                                                        *
 *************************************************************************/

// https://gist.github.com/rlipscombe/0c0f6b6057f398df4e36

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

#define MSS_TCP 500
#define CLIENT 0
#define SERVER 1
#define PORT 55555

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
  fprintf(stderr, "%s [-p <port>] [-d]\n", progname);
  fprintf(stderr, "%s -h\n", progname);
  fprintf(stderr, "\n");
  fprintf(stderr, "-p <port>: port to listen on, default 55555\n");
  fprintf(stderr, "-d: outputs debug information while running\n");
  fprintf(stderr, "-h: prints this help text\n");
  exit(1);
}

int main(int argc, char *argv[]) {

  int option, nBytesRead = 1;
  uint16_t nread, nwrite, plength;
  char buffer[MSS_TCP];
  struct sockaddr_in local, remote;
  char remote_ip[16] = "";            /* dotted quad IP string */
  unsigned short int port = PORT;
  int net_fd, sock_fd, optval = 1;
  socklen_t remotelen;
  int cliserv = -1;    /* must be specified on cmd line */

  progname = argv[0];

  /* Check command line options */
  while((option = getopt(argc, argv, "p:dh")) > 0) {
    switch(option) {
      case 'd':
        debug = 1;
        break;
      case 'h':
        usage();
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

  if ( (sock_fd = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
    perror("socket()");
    exit(1);
  }

  /* Server, wait for connections */

  /* avoid EADDRINUSE error on bind() */
  if(setsockopt(sock_fd, SOL_SOCKET, SO_REUSEADDR, (char *)&optval, sizeof(optval)) < 0) {
    perror("setsockopt()");
    exit(1);
  }

  memset(&local, 0, sizeof(local));
  local.sin_family = AF_INET;
  local.sin_addr.s_addr = htonl(INADDR_ANY);
  local.sin_port = htons(port);
  if (bind(sock_fd, (struct sockaddr*) &local, sizeof(local)) < 0) {
    perror("bind()");
    exit(1);
  }

  if (listen(sock_fd, 5) < 0) {
    perror("listen()");
    exit(1);
  }

  /* wait for connection request */
  remotelen = sizeof(remote);
  memset(&remote, 0, remotelen);
  if ((net_fd = accept(sock_fd, (struct sockaddr*)&remote, &remotelen)) < 0) {
    perror("accept()");
    exit(1);
  }

  do_debug("SERVER: Client connected from %s\n", inet_ntoa(remote.sin_addr));

  while(1) {

    /* Read length */
    // nread = read_n(net_fd, (char *)&plength, sizeof(plength));
    // if(nread == 0) {
      /* ctrl-c at the other end */
    //  break;
    //}

    /* read packet */
    nread = read_n(net_fd, buffer, sizeof(buffer));

    /*
    if(nread > 0)
    {
      do_debug("SERVER: Read %d bytes.\n", nread);
      for(int i = 0; i < BUFSIZE; i++)
      {
        printf("%d ", buffer[i]);
      }
      printf("\n");
    }
    */
  }

  return(0);
}
