# CIMA Ensemble Runner

!OK * aggiornare script per eseguire tutti gli ensemble member
!OK * moltiplicare il post process su tutti gli ensemble member
!OK * implementare il post process

* testare malfunzionamenti
OK * aggiungere verifica successo a da e wrf
OK * fare avgvar solo se >= 24 ore
OK * aggiungere verifica successo dei vari processi
OK * implementare opzione AssimilateOnlyInnerDomain
OK * implementare opzione AssimilateFirstCycle
OK * rimuovere creazione AUX dom. 1 e 2
OK * rimuovere creazione wrfout, lasciare solo AUX dom. 3
OK * modificare percorsi da cui leggere le obs
OK * rendere possibile saltare la parte WPS
OK * assicurarsi di linkare da obs anche le wunder
NO * inserire output nei log che visualizzi il progresso nella creazione dei files AUX?
OK * confrontare i template namelist attuali con quelli di LEXIS 
OK * aggiornare il template wrf00 per includere proprietá nalla namelist per ensemble. usare un nuovo template dir, perché quello attuale va usato per la run di controllo
OK * confrontare i numeri di cores in cfg fra attuali e quelli di LEXIS 
OK * aggiungere assimilazione su tutti i domini
OK * aggiungere assimilazione nel cicli delle 18
OK * nei log, indicare le dir relative a workdir/DATA. indicare all'inizio del log la directory WORKDIR
OK * rename WRFITA_ROOTDIR to ROOTDIR
OK * creare $WORKDIR come workdir/YYYY-MM-DD-HH ed usare nei logs


DELIVERIES OL 

continuum - aux 3 non regridded
dewetra - aux 3 regridded+merged
dewetra - aux 2 regridded+merged

arpal - out 3 upped
repo - out 3 upped + phases
vda - out 3 upped



# Consegna su repository AWS per ETT solutions
AWS_SRV=wrfrepo@34.245.16.170 
OL_PATH=repository/ol
ssh -i /home/antonio/.ssh/id_rsa.wrfprod $AWS_SRV "rm -rf $OL_PATH; mkdir -p $OL_PATH"
scp -i /home/antonio/.ssh/id_rsa.wrfprod rg_wrf-${RUNDATE}_00UTC.nc $AWS_SRV:$OL_PATH/rg_wrf-${RUNDATE}_00UTC.nc.tmp
ssh -i /home/antonio/.ssh/id_rsa.wrfprod $AWS_SRV mv $OL_PATH/rg_wrf-${RUNDATE}_00UTC.nc.tmp $OL_PATH/rg_wrf-${RUNDATE}_00UTC.nc

