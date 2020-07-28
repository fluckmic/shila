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
#define MB_TO_WRITE 25
#define BYTES_PER_MB 1000000
#define MSS_TCP 512

#define BUFFER_SIZE 1000000  // 0.1 MByte

#define CLIENT 0
#define SERVER 1

#define DEBUG 0

char *progname;

/**************************************************************************
 * do_debug: prints debugging stuff                                       *
 **************************************************************************/
void do_debug(char *msg, ...)
{

  va_list argp;

  if(DEBUG)
  {
	va_start(argp, msg);
	vfprintf(stderr, msg, argp);
	va_end(argp);
  }
}

/**************************************************************************
 * my_err: prints custom error messages on stderr.                        *
 **************************************************************************/
void my_err(char *msg, ...)
{

  va_list argp;

  va_start(argp, msg);
  vfprintf(stderr, msg, argp);
  va_end(argp);
}

/**************************************************************************
 * cread: read routine that checks for errors and exits if an error is    *
 *        returned.                                                       *
 **************************************************************************/
int cread(int fd, char *buf, int n)
{

  int nread;

  if((nread=read(fd, buf, n)) < 0)
  {
    perror("Reading data");
    exit(1);
  }
  return nread;
}

/**************************************************************************
 * read_n: ensures we read exactly n bytes, and puts them into "buf".     *
 *         (unless EOF, of course)                                        *
 **************************************************************************/
int read_n(int fd, char *buf, int n)
{

  int nread, left = n;

  while(left > 0)
  {
    if ((nread = cread(fd, buf, left)) == 0)
    {
      return 0 ;
    }
    else
    {
      left -= nread;
      buf += nread;
    }
  }
  return n;
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
void usage(void)
{
  fprintf(stderr, "Usage:\n");
  fprintf(stderr, "%s [-s|-c <serverIP>] [-p <port>] [-n <MBytes to write>] [-r] \n", progname);
  fprintf(stderr, "%s -h\n", progname);
  fprintf(stderr, "\n");
  fprintf(stderr, "-s|-c <serverIP>: run in server mode (-s), or specify server address (-c <serverIP>) (mandatory)\n");
  fprintf(stderr, "-p <port>: port to listen on (if run in server mode) or to connect to (in client mode), default 55555\n");
  fprintf(stderr, "-r: Entity is the receiving side.\n");
  fprintf(stderr, "-n <MBytes to write>: number of mega bytes to send, default 25\n");
  fprintf(stderr, "-h: prints this help text\n");
  exit(1);
}

int main(int argc, char *argv[])
{

  int isReceiver = 0;
  int option;
  int mbToWrite = MB_TO_WRITE;
  int sockFd;
  int netFd;
  int optionValue = 1;
  int clientOrServer = -1;

  socklen_t remoteLen;

  char remoteIP[16] = "";

  struct sockaddr_in remoteAddr;
  struct sockaddr_in localAddr;

  unsigned short int port = PORT;

  char buffer[BUFFER_SIZE];

  struct timespec timeSendingStart;
  struct timespec timeSendingEnd;

  progname = argv[0];

  /* Check command line options */
  while((option = getopt(argc, argv, "sc:p:n:r")) > 0)
  {
    switch(option)
    {
      case 'h':
        usage();
        break;
      case 'c':
        clientOrServer = CLIENT;
        strncpy(remoteIP,optarg,15);
        break;
      case 's':
        clientOrServer = SERVER;
        break;
      case 'n':
        mbToWrite = atoi(optarg);
        break;
      case 'p':
        port = atoi(optarg);
        break;
      case 'r':
        isReceiver = 1;
        break;
      default:
        my_err("Unknown option %c\n", option);
        usage();
    }
  }

  argv += optind;
  argc -= optind;

  if(argc > 0)
  {
    my_err("Too many options!\n");
    usage();
  }

  if(clientOrServer < 0)
  {
      my_err("Must specify client or server mode!\n");
      usage();
  }
  else if((clientOrServer == CLIENT) && (*remoteIP == '\0'))
  {
        my_err("Must specify server address!\n");
        usage();
  }

  if ( (sockFd = socket(AF_INET, SOCK_STREAM, 0)) < 0)
  {
    perror("socket()");
    exit(1);
  }

  if(clientOrServer == CLIENT)
  {
        /* set the maximum segment size */
        optionValue = MSS_TCP;
        if(setsockopt(sockFd, IPPROTO_TCP, TCP_MAXSEG, (char *)&optionValue, sizeof(optionValue)) < 0)
        {
          perror("setsockopt()");
          exit(1);
        }

        /* assign the destination address */
        memset(&remoteAddr, 0, sizeof(remoteAddr));
        remoteAddr.sin_family = AF_INET;
        remoteAddr.sin_addr.s_addr = inet_addr(remoteIP);
        remoteAddr.sin_port = htons(port);

        /* connection request */
        if (connect(sockFd, (struct sockaddr*) &remoteAddr, sizeof(remoteAddr)) < 0)
        {
        perror("connect()");
        exit(1);
        }

        netFd = sockFd;
        do_debug("Client: Connected to server %s\n", inet_ntoa(remoteAddr.sin_addr));
  }
  else
  {
        /* Server, wait for connections */
        /* avoid EADDRINUSE error on bind() */
        if(setsockopt(sockFd, SOL_SOCKET, SO_REUSEADDR, (char *)&optionValue, sizeof(optionValue)) < 0)
        {
          perror("setsockopt()");
          exit(1);
        }

        memset(&localAddr, 0, sizeof(localAddr));
        localAddr.sin_family = AF_INET;
        localAddr.sin_addr.s_addr = htonl(INADDR_ANY);
        localAddr.sin_port = htons(port);
        if (bind(sockFd, (struct sockaddr*) &localAddr, sizeof(localAddr)) < 0)
        {
          perror("bind()");
          exit(1);
        }

        if (listen(sockFd, 5) < 0)
        {
          perror("listen()");
          exit(1);
        }

        do_debug("Server: Listening on port %d.\n", port);

        /* wait for connection request */
        remoteLen = sizeof(remoteAddr);
        memset(&remoteAddr, 0, remoteLen);
        if ((netFd = accept(sockFd, (struct sockaddr*)&remoteAddr, &remoteLen)) < 0)
        {
          perror("accept()");
          exit(1);
        }

        do_debug("Server: Client connected from %s\n", inet_ntoa(remoteAddr.sin_addr));
  }

  if(isReceiver)
  {
        int nBytesReadTotal = 0;
        int nMBytesReceived = 0;
        int nBytesRead;
        while(1)
        {
            nBytesRead = read_n(netFd, buffer, sizeof(buffer));
            /*
            if(nBytesRead > 0)
            {
                nBytesReadTotal += nBytesRead;
                if(nBytesReadTotal % BYTES_PER_MB == 0 && nBytesReadTotal > 0)
                 {
                    nMBytesReceived++;
                    do_debug("Receiver: Received %d MBytes so far.\n",  nMBytesReceived);
                 }
                 if(nMBytesReceived >= mbToWrite)
                 {
                    do_debug("Receiver: Received the required amount of %d MBytes.\n", mbToWrite);
                    exit(0);
                 }
            }
            */
        }
  }
  else
  {
        for(int i = 0; i < BUFFER_SIZE; i++)
        {
            buffer[i] = i % 256;
        }

        int nBytesWritten = 0;
        int nBytesWrittenTotal = 0;
        int nMBytesWritten = 0;

        printf("Start sending %d MBytes..\n", mbToWrite);
        get_now(&timeSendingStart);

        while(nMBytesWritten < mbToWrite)
        {
            nBytesWritten = write(netFd, buffer, sizeof(buffer));
            if( nBytesWritten < 0)
            {
                perror("Writing data");
                exit(1);
            }
            nBytesWrittenTotal += nBytesWritten;
            if(nBytesWrittenTotal % BYTES_PER_MB == 0 && nBytesWrittenTotal > 0)
             {
                  nMBytesWritten++;
                  do_debug("Sender: Sent %d MBytes so far.\n",  nMBytesWritten);
             }
        }

        get_now(&timeSendingEnd);

        close(netFd);

        printf("%ld, %ld, %ld, %ld, %d\n",
            timeSendingStart.tv_sec, timeSendingStart.tv_nsec,
            timeSendingEnd.tv_sec, timeSendingEnd.tv_nsec,
            nMBytesWritten);

        float deltaTime = (timeSendingEnd.tv_sec - timeSendingStart.tv_sec) +
            ((timeSendingEnd.tv_nsec - timeSendingStart.tv_nsec) / 1e9);

        printf("%f s\n", deltaTime);

        float goodput = nMBytesWritten / deltaTime;

        printf("%f MB/s\n", goodput);
  }

  return(0);
}
