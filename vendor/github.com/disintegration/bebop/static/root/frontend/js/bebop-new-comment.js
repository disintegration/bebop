var BebopNewComment = Vue.component("bebop-new-comment", {
  template: `
    <div class="container content-container">
      <h2>New Comment</h2>
      <div>
        <div class="form-group">
          <label for="user-name" class="form-control-label">Comment:</label>
          <textarea class="form-control" id="comment-input" @change="hideErrorMessage" @keyup="hideErrorMessage" maxlength="10000"></textarea>
        </div>
        <div id="form-error" class="alert alert-danger" :class="{hidden: errorMessage===''}" role="alert" style="cursor:pointer" @click="hideErrorMessage">
          {{errorMessage}}
        </div>
      </div>
      <div>
        <button type="button" class="btn btn-primary btn-sm" @click="postComment" :disabled="posting">
          <i class="fa fa-reply"></i> Reply
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
    postComment: function() {
      var topicId = parseInt(this.$route.params.topic, 10);
      var comment = $("#comment-input").val().trim();
      if (comment.length < 1 || comment.length > 10000) {
        this.showErrorMessage("Invalid comment");
        return;
      }
      this.posting = true;
      this.$http
        .post("api/v1/comments", {
          topic: topicId,
          content: comment,
        })
        .then(
          response => {
            var id = response.data.id;
            var page = Math.floor((response.data.count - 1) / COMMENTS_PER_PAGE) + 1;
            this.posting = false;
            this.$parent.$router.push("/t/" + topicId + "/p/" + page + /c/ + id);
          },
          response => {
            this.posting = false;
            this.showErrorMessage("An error occured");
            console.log("ERROR: postComment: " + JSON.stringify(response.body));
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
