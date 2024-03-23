document.addEventListener('alpine:init', () => {
  Alpine.store('state', {
    init: function () {
      this.fetchLogs();
    },
    live: false,
    data: [],
    currentLog: {},
    fetchLogs: function () {
      fetch('/logs')
        .then((response) => response.json())
        .then((data) => {
          this.data = data;
        });
    },
    addToLogs: function (log) {
      this.data.unshift(log);
    },
    toggleLive: function () {
      this.live = !this.live;
    },
    selectLog: function (id) {
      this.currentLog = this.data.find((log) => log.id === id);
    },
  });
});
