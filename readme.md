# CIMA Ensemble Runner

OK * fare avgvar solo se >= 24 ore
* modificare percorsi da cui leggere le obs
* rendere possibile saltare la parte WPS
* assicurarsi di linkare da obs anche le wunder
* implementare opzione AssimilateOnlyInnerDomain
* 
* confrontare i template namelist attuali con quelli di LEXIS 
* aggiornare il template wrf00 per includere proprietá nalla namelist per ensemble. usare un nuovo template dir, perché quello attuale va usato per la run di controllo
* aggiornare script per eseguire tutti gli ensemble member
* testare malfunzionamenti
* moltiplicare il post process su tutti gli ensemble member
* confrontare i numeri di cores in cfg fra attuali e quelli di LEXIS 
* inserire output nei log che visualizzi il progresso nella creazione dei files AUX?
* rimuovere creazione wrfout, lasciare solo AUX dom. 3
OK * aggiungere assimilazione su tutti i domini
OK * aggiungere assimilazione nel cicli delle 18
OK * nei log, indicare le dir relative a workdir/DATA. indicare all'inizio del log la directory WORKDIR
OK * rename WRFITA_ROOTDIR to ROOTDIR
OK * creare $WORKDIR come workdir/YYYY-MM-DD-HH ed usare nei logs
