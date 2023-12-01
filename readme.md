# CIMA Ensemble Runner

* confrontare i template namelist attuali con quelli di LEXIS 
* aggiornare il template wrf00 per includere proprietá nalla namelist per ensemble. usare un nuovo template dir, perché quello attuale va usato per la run di controllo
* aggiornare script per eseguire tutti gli ensemble member
* modificare percorsi da cui leggere le obs
OK * aggiungere assimilazione su tutti i domini
OK * aggiungere assimilazione nel cicli delle 18
* fare avgvar solo se >= 24 ore
* testare malfunzionamenti
* moltiplicare il post process su tutti gli ensemble member
OK * nei log, indicare le dir relative a workdir/DATA. indicare all'inizio del log la directory WORKDIR
* confrontare i numeri di cores in cfg fra attuali e quelli di LEXIS 
* inserire output nei log che visualizzi il progresso nella creazione dei files AUX?
* rimuovere creazione wrfout, lasciare solo AUX dom. 3
OK * rename WRFITA_ROOTDIR to ROOTDIR
OK * creare $WORKDIR come workdir/YYYY-MM-DD-HH ed usare nei logs
* rendere possibile saltare la parte WPS