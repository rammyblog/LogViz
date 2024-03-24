document.addEventListener('alpine:init', () => {
  Alpine.store('state', {
    init: function () {
      this.fetchLogs();
    },
    live: false,
    data: [],
    searchBy: 'Search By',
    methodType: 'Select Method',
    isMethodTypeOpen: false,
    isLoading: false,
    currentLog: {},
    fetchLogs: function () {
      fetch('/logs?lastId=0')
        .then((response) => response.json())
        .then((data) => {
          this.data = data;
        });
    },
    addToLogs: function (log) {
      this.data.unshift(log);
    },

    searchLogs: function (searchTerm) {
      console.log(searchTerm)
      const validSearchBy = ['Path', 'Code', 'Ip Address', 'Method'];
      this.isLoading = true;
      if (!validSearchBy.includes(this.searchBy)) {
        this.isLoading = false;
        return;
      }
      fetch(
        '/logs?searchBy=' +
          this.searchBy +
          '&searchTerm=' +
          searchTerm +
          '&lastId=0'
      )
        .then((response) => response.json())
        .then((data) => {
          this.data = data;
          this.isLoading = false;
        });
    },

    loadMore: function () {
      this.isLoading = true;
      const lastItem = this.data[this.data.length - 1];
      const lastId = lastItem ? lastItem.id : '0';
      fetch('/logs?lastId=' + lastId)
        .then((response) => response.json())
        .then((data) => {
          if (data.length === 0) {
            this.isLoading = false;
            return;
          }
          this.data = this.data.concat(data);
          this.isLoading = false;
        });
    },
    toggleLive: function () {
      this.live = !this.live;
    },
    selectLog: function (id) {
      this.currentLog = this.data.find((log) => log.id === id);
    },
    toggleSearchBy: function (searchBy) {
      if (searchBy === 'Method') {
        this.isMethodTypeOpen = !this.isMethodTypeOpen;
        this.searchBy = searchBy;
        return;
      }
      this.isMethodTypeOpen = false;
      this.searchBy = searchBy;
    },
    toggleMethodType: function (methodType) {
      this.methodType = methodType;
      this.searchLogs(methodType);
    },
  });
});
