setlocal EnableDelayedExpansion
go build

FOR %%x IN (1 2 3) DO (
    FOR /L %%y IN (1 1 1000) DO (
        del simdata.csv
        .\dowts.exe %%x 
        ::mongoexport --collection=logs --db=dowsimlogs --out=simdata.csv --type=csv --fields="IP,bot,functioID,timestamp"
        ::mongosh dowsimlogs --eval "db.logs.drop()"
        ::move simdata.csv D:\DoWDetect
        py -3 C:\Users\daniel\Documents\DoWDetect\heatmap.py %%y %%x
        
)
)
