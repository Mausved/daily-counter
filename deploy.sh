#!bin/bash

session="daily_counter"
pid=$(ps -e | grep $session | grep -v 'grep' | awk '{print $1}')
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

tmux new-session -d -s $session "go build -o $session main.go processor.go && ./$session"
