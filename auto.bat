setlocal EnableDelayedExpansion
go build
.\dowts.exe
mongoexport --collection=logs --db=dowsimlogs --out=simdata.csv --type=csv --fields="IP,bot,functioID,timestamp"
mongo dowsimlogs --eval "db.logs.drop()"
::move simdata.csv D:\DoWDetect
::py -3 D:\DoWAnalyser\DoWAnalyser\graphwork.py