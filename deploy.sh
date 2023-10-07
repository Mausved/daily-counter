#!bin/bash

pid=$(ps | grep '\./main' | grep -v 'grep' | awk '{print $1}')
if [[ $pid -ne '' ]]
then
    kill -2 $pid
fi

cd dailycounter && tmux new-session -d -s daily_counter 'go build -o main main.go processor.go && ./main'