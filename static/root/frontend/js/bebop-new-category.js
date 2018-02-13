var BebopNewCategory = Vue.component("bebop-new-category", {
  template: `
    <div class="container content-container">
      <h2>New Category</h2>
      <div>
        <div class="form-group">
          <label for="user-name" class="form-control-label">Title:</label>
          <input type="text" class="form-control" id="category-title-input" @change="hideErrorMessage" @keyup="hideErrorMessage" maxlength="100">
        </div>
        <div id="form-error" class="alert alert-danger" :class="{hidden: errorMessage===''}" role="alert" style="cursor:pointer" @click="hideErrorMessage">
          {{errorMessage}}
        </div>
      </div>
      <div>
        <button type="button" class="btn btn-primary btn-sm" @click="postCategory" :disabled="posting">
          <i class="fa fa-plus"></i> Create Category
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
    postCategory: function() {
      var title = $("#category-title-input").val().trim();
      if (title.length < 1 || title.length > 100) {
        this.showErrorMessage("Invalid category title");
        return;
      }
      this.posting = true;
      this.$http
        .post("api/v1/categories", {
          title: title,
        })
        .then(
          response => {
            this.posting = false;
            this.$parent.$router.push("/c/" + response.data.id);
          },
          response => {
            this.posting = false;
            this.showErrorMessage("An error occured");
            console.log("ERROR: postCategory: " + JSON.stringify(response.body));
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

