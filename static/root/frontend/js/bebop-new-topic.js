var BebopNewTopic = Vue.component("bebop-new-topic", {
  template: `
    <div class="container content-container">
      <h2>New Topic</h2>
      <div>
        <div class="form-group">
          <label for="user-name" class="form-control-label">Title:</label>
          <input type="text" class="form-control" id="topic-title-input" @change="hideErrorMessage" @keyup="hideErrorMessage" maxlength="100">
        </div>
        <div class="form-group">
          <label for="user-name" class="form-control-label">Comment:</label>
          <textarea class="form-control" id="comment-input" @change="hideErrorMessage" @keyup="hideErrorMessage" maxlength="10000"></textarea>
        </div>
        <div id="form-error" class="alert alert-danger" :class="{hidden: errorMessage===''}" role="alert" style="cursor:pointer" @click="hideErrorMessage">
          {{errorMessage}}
        </div>
      </div>
      <div>
        <button type="button" class="btn btn-primary btn-sm" @click="postTopic" :disabled="posting">
          <i class="fa fa-plus"></i> Create Topic
        </button>
      </div>
    </div>
  `,

  props: ["config", "auth"],

  data: function() {
    return {
      errorMessage: "",
      posting: false,
    };
  },

  mounted: function() {
    $("#comment-input").markdown({
      iconlibrary: "fa",
      fullscreen: {
        enable: false,
      },
    });
  },

  methods: {
    postTopic: function() {
      var title = $("#topic-title-input").val().trim();
      if (title.length < 1 || title.length > 100) {
        this.showErrorMessage("Invalid topic title");
        return;
      }
      var comment = $("#comment-input").val().trim();
      if (comment.length < 1 || comment.length > 10000) {
        this.showErrorMessage("Invalid comment");
        return;
      }
      this.posting = true;
      this.$http
        .post("api/v1/topics", {
          title: title,
          content: comment,
        })
        .then(
          response => {
            this.posting = false;
            this.$parent.$router.push("/t/" + response.data.id);
          },
          response => {
            this.posting = false;
            this.showErrorMessage("An error occured");
            console.log("ERROR: postTopic: " + JSON.stringify(response.body));
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
