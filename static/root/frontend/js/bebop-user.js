var BebopUser = Vue.component("bebop-user", {
  template: `
    <div class="container content-container">

      <div v-if="!dataReady" class="loading-info">
        <div v-if="error" >
          <p class="text-danger">
            Sorry, could not load the user profile. Please check your connection.
          </p>
          <a class="btn btn-primary btn-sm" role="button" @click="load">
            <i class="fa fa-refresh"></i> Try Again
          </a>
        </div>
        <div v-else>
          <i class="fa fa-circle-o-notch fa-spin fa-3x fa-fw"></i>
        </div>
      </div>
      <div v-else>

        <h2 v-if="isMe">My profile</h2>
        <h2 v-else>User profile: {{user.name}}</h2>

        <div class="card user-profile">

          <div class="row">
            <div class="col-xs-3">
              Username
            </div>
            <div class="col-xs-6">
              {{user.name}}
            </div>
            <div class="col-xs-3 text-right">
              <label class="btn btn-fix" :class="{'btn-default': isMe, 'btn-danger': !isMe}" role="button" @click="changeUsername()"><i class="fa fa-pencil-square-o" aria-hidden="true"></i></label>
            </div>
          </div>

          <hr>

          <div class="row">
            <div class="col-xs-3">
              Avatar
            </div>
            <div class="col-xs-6">
              <div v-if="uploadingAvatar">
                <i class="fa fa-circle-o-notch fa-spin fa-2x fa-fw"></i>
              </div>
              <div v-else>
                <img v-if="user.avatar" class="img-circle" :src="user.avatar" width="35" height="35"> 
                <img v-else class="img-circle" src="data:image/gif;base64,R0lGODlhAQABAIAAAP///wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw==" width="35" height="35"> 
              </div>
            </div>
            <div class="col-xs-3 text-right">
              <label for="avatar-upload-input" class="btn btn-fix" :class="{'btn-default': isMe, 'btn-danger': !isMe}" role="button">
                <i class="fa fa-cloud-upload"></i>
              </label>
              <input id="avatar-upload-input" class="hidden" type="file" @change="uploadAvatar()"/>
            </div>
          </div>
          <div v-if="avatarUploadError" class="row">
            <div class="col-xs-12">
              <div class="alert alert-danger" style="margin-top:10px">{{avatarUploadError}}</div>
            </div>
          </div>

          <hr>

          <div class="row">
            <div class="col-xs-3">
              Sign in with
            </div>
            <div class="col-xs-6">
              {{user.authService|capitalize}}
            </div>
          </div>

          <hr>

          <div class="row">
            <div class="col-xs-3">
              Activated
            </div>
            <div class="col-xs-6">
              {{user.createdAt|formatTime}}
            </div>
          </div>
          
          <hr v-if="auth.authenticated && auth.user.admin && !isMe">

          <div v-if="auth.authenticated && auth.user.admin && !isMe" class="row">
            <div class="col-xs-3">
              Blocked
            </div>
            <div class="col-xs-6">
              <span v-if="user.blocked" class="text-danger">Yes</span>
              <span v-else class="text-success">No</span>
            </div>
            <div class="col-xs-3 text-right">
              <button v-if="user.blocked" class="btn btn-danger btn-fix" @click="setBlocked(false)"><i class="fa fa-unlock-alt" aria-hidden="true"></i></button>
              <button v-else class="btn btn-danger btn-fix" @click="setBlocked(true)"><i class="fa fa-lock" aria-hidden="true"></i></button>
            </div>
          </div>

        </div>
      </div>

    </div>
  `,

  props: ["config", "auth"],

  data: function() {
    return {
      user: {},
      userReady: false,
      error: false,
      uploadingAvatar: false,
      avatarUploadError: "",
    };
  },

  computed: {
    dataReady: function() {
      return this.userReady;
    },

    userId: function() {
      if (!this.auth.authenticated) {
        return 0;
      }

      var userId = parseInt(this.$route.params.user, 10);
      if (!userId) {
        return this.auth.user.id;
      }

      return userId;
    },

    isMe: function() {
      if (!this.auth.authenticated) {
        return false;
      }
      return this.userId === this.auth.user.id;
    },
  },

  watch: {
    userId: function(val) {
      this.load();
    },
  },

  created: function() {
    this.load();
  },

  methods: {
    load: function() {
      this.user = {};
      this.userReady = false;
      this.error = false;
      this.uploadingAvatar = false;
      this.avatarUploadError = "";
      this.getUser();
    },

    getUser: function() {
      if (!this.auth.authenticated) {
        this.$parent.$router.replace("/");
        return;
      }

      if (!this.auth.user.admin && this.auth.user.id !== this.userId) {
        this.$parent.$router.replace("/me");
        return;
      }

      var url = "api/v1/me";
      if (this.auth.user.id !== this.userId) {
        url = "api/v1/users/" + this.userId;
      }

      this.$http.get(url).then(
        response => {
          this.user = response.body.user;
          this.userReady = true;
        },
        response => {
          console.log("ERROR: getUser: " + JSON.stringify(response.body));
          this.error = true;
        }
      );
    },

    changeUsername: function() {
      if (!this.userReady) {
        return;
      }
      this.$parent.$refs.usernameModal.show(this.userId, this.user.name, success => {
        if (success) {
          if (this.isMe) {
            this.$parent.getMe();
          }
          this.load();
        }
      });
    },

    uploadAvatar: function() {
      var input = document.getElementById("avatar-upload-input");
      var file = input.files[0];
      input.value = "";
      var reader = new FileReader();
      reader.onload = () => {
        var parts = reader.result.split(";base64,");
        var imageData = "";
        if (parts.length === 2) {
          imageData = parts[1];
        }
        this.putUserAvatar(imageData);
      };
      reader.readAsDataURL(file);
    },

    putUserAvatar: function(imageData) {
      if (!this.userReady) {
        return;
      }
      this.uploadingAvatar = true;
      this.avatarUploadError = "";
      this.$http.put("api/v1/users/" + this.userId + "/avatar", { avatar: imageData }).then(
        response => {
          if (this.isMe) {
            this.$parent.getMe();
          }
          this.uploadingAvatar = false;
          this.load();
        },
        response => {
          console.log("ERROR: putUserAvatar: " + JSON.stringify(response.body));
          this.uploadingAvatar = false;
          var error = "Sorry, could not upload that image. An error occured.";
          if (response.body.error && response.body.error.code === "BadRequest") {
            error = "Sorry, could not upload that image. ";
            error += "Please choose an image from 50x50 to 2000x2000 pixels in size. ";
            error += "The supported formats are JPEG, PNG, GIF, TIFF, BMP. ";
            error += "The maximum file size is 5MB.";
          }
          this.avatarUploadError = error;
        }
      );
    },

    setBlocked(val) {
      action = val ? "block" : "unblock";
      if (!confirm("Are you sure you want to " + action + " this user?")) {
        return;
      }
      this.$http.put("api/v1/users/" + this.userId + "/blocked", { blocked: val }).then(
        response => {
          this.load();
        },
        response => {
          console.log("ERROR: setBlocked: " + JSON.stringify(response.body));
        }
      );
    },
  },
});
