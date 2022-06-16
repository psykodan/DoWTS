setlocal EnableDelayedExpansion
go build
.\dowts.exe
mongoexport --collection=logs --db=dowsimlogs --out=simdata.csv --type=csv --fields="IP,bot,functioID,timestamp"
mongo dowsimlogs --eval "db.logs.drop()"
move simdata.csv C:\Users\daniel\Documents\PhD\Papers\DoWTS_Denial_of_Wallet_Test_Simulator_Modelling_Attack_Vectors_for_Preemptive_Defence\review\data
py -3 D:\DoWAnalyser\DoWAnalyser\graphwork.py