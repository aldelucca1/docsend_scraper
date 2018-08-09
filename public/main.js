var connection = null;
new Vue({
  el: '#app',
  data: {
    email: null,
    documents: []
  },
  methods: {
    login(evt) {
      var form = $(evt.target);
      this.email = $("input", form).val();
      this.connectSocket();
      this.getData();
    },
    logout() {
      this.email = null;
    },
    connectSocket() {
      connection = new WebSocket('ws://' + window.location.hostname + ':8080/api/status?owner=' + this.email);
      connection.onopen = () => {
        console.log('WebSocket connected');
      };

      // Log errors
      connection.onerror = (error) => {
        console.log('WebSocket Error ' + error);
      };

      // Log messages from the server
      connection.onmessage = (e) => {
        var message = JSON.parse(e.data);
        if (message.type == "PING") {
          connection.send({type: "PONG"});
        } else {
          var document = message.data;
          for (let i = 0, n = this.documents.length; i < n; i++) {
            if (this.documents[i].id == document.id) {
              this.documents.splice(i, 1, document);
              break;
            }
          }
        }
      };
    },
    getData() {
      var route = '/api/documents?owner=' + this.email;
      $.get(route, res => {
        this.documents = res;
      });
    },
    generate(evt) {
      var route = '/api/documents';
      var form = $(evt.target);
      var sourceURL = $("input[name=source_url]", form);
      if (!/^https\:\/\/docsend\.com\/view\/.*$/.test(sourceURL.val())) {
        sourceURL.parent().addClass("has-error");
        return
      }
      $.post(route, form.serialize(), res => {
        this.documents.unshift(res);
      })
      $('#generate-modal').modal('hide');
    },
    download(id) {
      window.location = '/api/documents/' + id + '/download';
    }
  },
  mounted() {
    $('#generate-modal').on('hidden.bs.modal', function (e) {
      var form = $("form", e.target);
      form[0].reset();
      $(".form-group", form).removeClass("has-error");
    });
  }
})
