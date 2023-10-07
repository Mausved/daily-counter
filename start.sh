appPid=$(ps | grep "main.go" | grep -v "grep" | awk '{print $1}')
if [[ $pid -eq '' ]]
then
    cd /home/mausved/dailycounter/ && go run main.go processor.go
fi