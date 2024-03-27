document.addEventListener('alpine:init', () => {
  Alpine.store('state', {
    init: function () {
      this.fetchLogs();
    },
    live: false,
    data: [],
    searchBy: 'Search By',
    searchTerm: '',
    methodType: 'Select Method',
    isMethodTypeOpen: false,
    isLoading: false,
    currentLog: {},
    lastId: null,
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
    validSearchBy: ['Path', 'Code', 'Ip Address', 'Method'],

    searchLogs: function (searchTerm) {
      this.isLoading = true;
      if (!this.validSearchBy.includes(this.searchBy)) {
        this.isLoading = false;
        return;
      }
      this.searchTerm = searchTerm;
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

      if (this.lastId === lastId) {
        this.isLoading = false;
        return;
      }
      this.lastId = lastId;

      fetch(
        '/logs?lastId=' +
          lastId +
          '&searchBy=' +
          this.searchBy +
          '&searchTerm=' +
          this.searchTerm
      )
        .then((response) => response.json())
        .then((data) => {
          if (data.length === 0) {
            this.isLoading = false;
            return;
          }
          console.log(data);
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
