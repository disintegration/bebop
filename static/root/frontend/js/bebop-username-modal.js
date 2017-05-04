var BebopUsernameModal = Vue.component("bebop-username-modal", {
  template: `
    <div class="modal fade" id="username-modal" tabindex="-1" role="dialog" data-backdrop="static">
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Username</h2>
          </div>
          <div class="modal-body">
            <div style="margin-bottom: 15px;">
              Please choose a username that is between 3 and 20 characters in length and containing only 
              alphanumeric characters (letters A-Z, numbers 0-9), hyphens, and underscores.
            </div>
            <div class="form-group">
              <label for="user-name" class="form-control-label">Username:</label>
              <input type="text" class="form-control" id="username-modal-input" v-model="name" @change="hideErrorMessage" @keyup="hideErrorMessage" @keyup.13="send">
            </div>
            <div id="username-modal-error" class="alert alert-danger" :class="{hidden: errorMessage===''}" role="alert" style="cursor:pointer" @click="hideErrorMessage">
              {{errorMessage}}
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-default" data-dismiss="modal">Cancel</button>
            <button type="button" class="btn btn-primary" id="username-modal-ok" @click="send">OK</button>
          </div>
        </div>
      </div>
    </div>
  `,

  data: function() {
    return {
      userId: 0,
      success: false,
      callback: function() {},
      name: "",
      errorMessage: "",
    };
  },

  mounted: function() {
    $("#username-modal").on("hidden.bs.modal", () => {
      this.callback(this.success);
    });
    $("#username-modal").on("shown.bs.modal", () => {
      $("#username-modal-input")[0].focus();
    });
  },

  methods: {
    show: function(userId, initialName, callback) {
      this.userId = userId;
      this.success = false;
      this.callback = callback;
      this.errorMessage = "";
      this.name = initialName;
      $("#username-modal").modal("show");
    },

    send: function() {
      this.$http.put("api/v1/users/" + this.userId + "/name", { name: this.name }).then(
        response => {
          this.success = true;
          $("#username-modal").modal("hide");
        },
        response => {
          if (response.data.error && response.data.error.code === "UnavailableUserName") {
            this.showErrorMessage("Sorry, that username is taken.");
          } else if (response.data.error && response.data.error.code === "InvalidUserName") {
            this.showErrorMessage("Invalid username.");
          } else {
            this.showErrorMessage("An error occured.");
          }
          $("#username-modal-input")[0].focus();
        }
      );
    },

    showErrorMessage: function(message) {
      this.errorMessage = message;
    },

    hideErrorMessage: function() {
      this.errorMessage = "";
    },
  },
});
