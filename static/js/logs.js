const logsApp = () => ({
  logs: [],
  init: function () {
    this.fetchLogs();
  },
  fetchLogs: function () {
    fetch('/api/logs')
      .then((response) => response.json())
      .then((data) => {
        this.logs = data;
      });
  },
});
