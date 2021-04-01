#!/bin/bash
FILE_NAME="cbsignal"

if [ -n "$2" ]; then
  SERVER=$2
else
  SERVER="signald"
fi

SERVICE="$SERVER.service"
PREFIX="/lib/systemd/system"
BASE_DIR=$PWD
INTERVAL=2

EXEC_COMMAND="$BASE_DIR/$SERVER"
if [ -n "$3" ]; then
  EXEC_COMMAND="$EXEC_COMMAND -c $3"
fi
echo "$EXEC_COMMAND"

# 命令行参数，需要手动指定
ARGS=""

function generateServiceFile()
{
echo "[Unit]
Description=cbsignal

[Service]
Type=simple
LimitCORE=infinity
LimitNOFILE=1000000
LimitNPROC=1000000


WorkingDirectory=$BASE_DIR

ExecStart=$EXEC_COMMAND

StandardOutput=null
StandardError=null

Restart=on-failure
ExecStop=

[Install]
WantedBy=multi-user.target" > ./$SERVICE
}

function deploy()
{
  if [ "$SERVER" != "$FILE_NAME" ]; then
    sudo cp $FILE_NAME $SERVER
  fi
  echo "Generate Service File"
	generateServiceFile
	sudo mv $SERVICE $PREFIX
	sudo systemctl daemon-reload
	sudo systemctl enable $SERVICE
}

function start()
{
  chmod +x $FILE_NAME

  sudo sysctl -w net.core.somaxconn=65535
  sudo sysctl -w fs.file-max=6000000
  sudo sysctl -w fs.nr_open=6000000
  ulimit -n 6000000

	deploy

	if [ "`pgrep $SERVER`" != "" ];then
		echo "$SERVER already running"
		exit 1
	fi

	sudo systemctl start $SERVICE
	echo "sleeping..." &&  sleep $INTERVAL

	# check status
	if [ "`pgrep $SERVER`" == "" ];then
		echo "$SERVER start failed"
		rm -f $SERVER
		exit 1
	fi

	echo "$SERVER start succeed"
}

function status()
{
	if [ "`pgrep $SERVER`" != "" ];then
		echo $SERVER is running
	else
		echo $SERVER is not running
	fi
}

function stop()
{
	if [ "`pgrep $SERVER`" != "" ];then
		sudo systemctl stop $SERVICE
		echo "sleeping..." &&  sleep $INTERVAL
		if [ "`pgrep $SERVER`" != "" ];then
		  echo "$SERVER stop failed"
		  exit 1
	  fi
	  rm -f $SERVER
	  echo "$SERVER stop succeed"
	else
	  echo "$SERVER is not running"
	fi
}

function restart()
{
	stop
  start
}

function test()
{
    if [ "`pgrep $SERVER -u $UID`" != "" ];then
        kill -9 `pgrep $SERVER -u $UID`
    fi

    echo "sleeping..." &&  sleep $INTERVAL

    if [ "`pgrep $SERVER -u $UID`" != "" ];then
        echo "$SERVER stop failed"
        exit 1
    fi

    echo "$SERVER stop succeed"

    ulimit -n 1000000

    nohup $BASE_DIR/$SERVER &>/dev/null &

    echo "sleeping..." &&  sleep $INTERVAL

    # check status
    if [ "`pgrep $SERVER -u $UID`" == "" ];then
        echo "$SERVER start failed"
        exit 1
    fi

    echo "$SERVER start succeed"
}

case "$1" in
	'start')
	start
	;;
	'stop')
	stop
	;;
	'status')
	status
	;;
	'restart')
	restart
	;;
	'test')
    test
    ;;
	*)
	echo "usage: $0 {start|stop|restart|status}"
	exit 1
	;;
esac
