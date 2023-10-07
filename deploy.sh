#!bin/bash

session="daily_counter"
env GOOS=linux GOARCH=amd64 go build -o $session main.go processor.go
rsync -avz ./$session  mausved@62.84.117.24:/home/mausved/dailycounter

restart() {
    session_name="daily_counter"
    pid=$(ps -e | grep ${session} | grep -v 'grep' | awk '{print $1}')
    echo "pid=$pid"

    if [[ $pid -ne '' ]]
    then
        kill -2 $pid
        wait $pid
        echo "stopped pid=$pid"
        tmux kill-session -t $session_name
        echo "killed session $session_name"
    else
        echo "not found already running app"
    fi
    tmux new-session -d -s $session_name "cd dailycounter && ./daily_counter"
    echo "started app"
}

ssh mausved@62.84.117.24 "$(typeset -f); restart"
rm $session
