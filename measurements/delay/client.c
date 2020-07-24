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
#include <time.h>
#include <errno.h>
#include <stdarg.h>
#include <netinet/tcp.h>

#define PORT 55555
#define MSS_TCP 500

#define MIN_TIME_BETWEEN_TWO_WRITES_NS 500
#define N_WRITES_DEFAULT 1000

FILE *logfile;

int debug;
char *progname;

/**************************************************************************
 * do_debug: prints debugging stuff                                       *
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
 * get_now: get current time                                              *
 **************************************************************************/
void get_now( struct timespec *time)
{
    if (clock_gettime(CLOCK_REALTIME, time) != 0 )
    {
      my_err("Can't get current time.\n");
    }
    return;
}

/**************************************************************************
 * usage: prints usage and exits.                                         *
 **************************************************************************/
void usage(void) {
  fprintf(stderr, "Usage:\n");
  fprintf(stderr, "%s [-c <serverIP>] [-n <nOfWrites>] [-p <port>] [-d] [-f <filename>] [-a <additionalLogLine>]\n", progname);
  fprintf(stderr, "%s -h\n", progname);
  fprintf(stderr, "\n");
  fprintf(stderr, "-c <serverIP>: server address (mandatory)\n");
  fprintf(stderr, "-f <filename>: name of the log file (mandatory)\n");
  fprintf(stderr, "-n <nOfWrites>: number of writes to do, default 1000\n");
  fprintf(stderr, "-a <additionalLogLine>: additional line which is written to the log file\n");
  fprintf(stderr, "-p <port>: port to connect to, default 55555\n");
  fprintf(stderr, "-d: outputs debug information while running\n");
  fprintf(stderr, "-h: prints this help text\n");
  exit(1);
}

int main(int argc, char *argv[]) {

  int option, nOfWrites = N_WRITES_DEFAULT;
  char buffer[MSS_TCP];
  struct sockaddr_in remote;
  char remote_ip[16] = "";            /* dotted quad IP string */
  unsigned short int port = PORT;
  int sock_fd, net_fd, optval = 1;
  struct timespec time_now;

  char *filename          = NULL;
  char *additionalLogLine = NULL;

  progname = argv[0];

  /* Check command line options */
  while((option = getopt(argc, argv, "c:p:f:n:a:hd")) > 0) {
    switch(option) {
      case 'a':
        additionalLogLine = optarg;
        break;
      case 'f':
        filename = optarg;
        break;
      case 'd':
        debug = 1;
        break;
      case 'h':
        usage();
        break;
      case 'c':
        strncpy(remote_ip,optarg,15);
        break;
      case 'n':
        nOfWrites = atoi(optarg);
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

  if(filename == NULL) {
    my_err("Must specify filename!\n");
    usage();
  }

  logfile = fopen( filename, "a+" );
  if ( logfile == NULL )
  {
    my_err("Could not open log file!\n");
  }

  if ( additionalLogLine != NULL )
  {
    fprintf(logfile,"%s\n", additionalLogLine);
      if ( fflush(logfile) != 0 )
      {
         perror("Writing log file");
         exit(1);
      }
  }

  if ( (sock_fd = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
    perror("socket()");
    exit(1);
  }

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

    do_debug("CLIENT: Try to connect to server %s\n", inet_ntoa(remote.sin_addr));

  /* connection request */
  if (connect(sock_fd, (struct sockaddr*) &remote, sizeof(remote)) < 0) {
    perror("connect()");
    exit(1);
  }

  net_fd = sock_fd;

  do_debug("CLIENT: Connected to server %s\n", inet_ntoa(remote.sin_addr));

  int A = 0; int B = 0; int C = 0; int D = 0; int E = 0;
  for(int writeCount = 0; writeCount < nOfWrites; writeCount++)
  {
    usleep(MIN_TIME_BETWEEN_TWO_WRITES_NS);

    A = writeCount % 256; if(A < 2) { A = 2; }
    B = B % 256; if (B < 2) { B = 2; }
    C = C % 256; if (C < 2) { C = 2; }
    D = D % 256; if (D < 2) { D = 2; }
    E = E % 256; if (E < 2) { E = 2; }

    for(int i = 0; i < MSS_TCP; i++)
    {
      if(i == MSS_TCP / 2)
      {
        buffer[i-3] = 1;
        buffer[i-2] = E;
        buffer[i-1] = D;
        buffer[i]   = C;
        buffer[i+1] = B;
        buffer[i+2] = A;
        i = i + 2;
      }
      else
      {
        buffer[i] = 0;
      }
    }

    get_now(&time_now);

    if(write(net_fd, buffer, sizeof(buffer)) < 0)
    {
      perror("Writing data");
      exit(1);
    }

    fprintf(logfile,"%ld, %ld, %d, %d, %d, %d, %d\n", time_now.tv_sec, time_now.tv_nsec, A, B, C, D, E);
    if ( fflush(logfile) != 0 )
    {
       perror("Writing log file");
       exit(1);
    }

    if(A == 255 && writeCount > 0) { B++; }
    if(B == 255) { C++; }
    if(C == 255) { D++; }
    if(D == 255) { E++; }

  }
  return(0);
}
