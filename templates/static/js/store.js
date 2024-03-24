document.addEventListener('alpine:init', () => {
  Alpine.store('state', {
    init: function () {
      this.fetchLogs();
    },
    live: false,
    data: [],
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
  });
});
