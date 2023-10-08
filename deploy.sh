#!bin/bash

session="daily_counter"
env GOOS=linux GOARCH=amd64 go build -o $session main.go processor.go
rsync -avz ./$session  $USER@$HOST:$PATH

restart() {
    session="daily_counter"
    pid=$(ps -e | grep ${session} | grep -v 'grep' | awk '{print $1}')
    echo "pid=$pid"

    if [[ $pid -ne '' ]]
    then
        kill -2 $pid
        wait $pid
        echo "stopped pid=$pid"
        tmux kill-session -t $session
        echo "killed session $session"
    else
        echo "not found already running app"
    fi
    tmux new-session -d -s $session "cd dailycounter && ./$session"
    echo "started app"
}

ssh $USER@$HOST "$(typeset -f); restart"
rm $session
