const CATEGORIES_PER_PAGE = 20;

var BebopCategories = Vue.component("bebop-categories", {
  template: `
    <div class="container content-container">

      <div v-if="!dataReady" class="loading-info">
        <div v-if="error" >
          <p class="text-danger">
            Sorry, could not load categories. Please check your connection.
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

        <div class="categories-category-top-buttons">
          <router-link v-if="auth.authenticated && auth.user.admin" to="/new-category" class="btn btn-primary btn-sm">
            <i class="fa fa-plus"></i> New Category
          </router-link>
          <a class="btn btn-primary btn-sm" role="button" @click="load">
            <i class="fa fa-refresh"></i> Refresh
          </a>
        </div>

        <h2 class="categories-title">Categories</h2>

        <nav v-if="page > 1">
          <ul class="pagination pagination-sm">
            <li v-for="p in pagination" :class="{active: page === p}">
              <span v-if="p === '...'">…</span>
              <router-link v-if="p !== '...'" :to="'/p/' + p">{{p}}</router-link>
            </li>
          </ul>
        </nav>

        <div v-for="category in categories" class="card categories-category">
          <div class="categories-category-title">
            <router-link :to="'/c/' + category.id">{{category.title}}</router-link>
          </div>
          <div class="categories-category-info">
                <span class="info-separator"> | </span>
                <i class="fa fa-comment-o"></i> {{category.topicCount}}
                <span class="info-separator"> | </span>
                <i class="fa fa-clock-o"></i> <span :title="category.lastTopicAt|formatTime">{{category.lastTopicAt|formatTimeAgo}}</span>
          </div>
          <div class="categories-category-admin-tools" v-if="auth.authenticated && auth.user.admin">
                <a class="a-tool" role="button" @click="delCategory(category.id)"><i class="fa fa-times" aria-hidden="true"></i> delete category</a>
                <span class="info-separator"> | </span> 
                <router-link :to="'/u/' + users[category.authorId].id" class="a-tool"><i class="fa fa-user" aria-hidden="true"></i> user profile</router-link>
          </div>
        </div>

        <nav v-if="lastPage > 1">
          <ul class="pagination pagination-sm">
            <li v-for="p in pagination" :class="{active: page === p}">
              <span v-if="p === '...'">…</span>
              <router-link v-if="p !== '...'" :to="'/p/' + p">{{p}}</router-link>
            </li>
          </ul>
        </nav>

      </div>

    </div>
  `,

  props: ["config", "auth"],

  data: function() {
    return {
      categories: [],
      categoriesReady: false,
      categoryCount: 0,
      users: {},
      usersReady: false,
      error: false,
    };
  },

  computed: {
    dataReady: function() {
      return this.categoriesReady && this.usersReady;
    },

    page: function() {
      var page = parseInt(this.$route.params.page, 10);
      if (!page || page < 1) {
        return 1;
      }
      return page;
    },

    lastPage: function() {
      if (!this.categoriesReady) {
        return 1;
      }
      var p = Math.floor((this.categoryCount - 1) / CATEGORIES_PER_PAGE) + 1;
      if (p < 1) {
        p = 1;
      }
      return p;
    },

    pagination: function() {
      if (!this.categoriesReady) {
        return [];
      }
      return getPagination(this.page, this.lastPage);
    },
  },

  watch: {
    page: function(val) {
      this.load();
    },
  },

  created: function() {
    this.load();
  },

  methods: {
    load: function() {
      this.categories = [];
      this.categoriesReady = false;
      this.categoryCount = 0;
      this.users = {};
      this.usersReady = false;
      this.error = false;
      this.getCategories();
    },

    getCategories: function() {
      var url = "api/v1/categories?limit=" + CATEGORIES_PER_PAGE;
      if (this.page > 1) {
        var offset = (this.page - 1) * CATEGORIES_PER_PAGE;
        url += "&offset=" + offset;
      }
      this.$http.get(url).then(
        response => {
          this.categories = response.body.categories;
          this.categoryCount = response.body.count;
          this.categoriesReady = true;

          if (this.page > this.lastPage) {
            this.$parent.$router.replace("/p/" + this.lastPage);
            return;
          }

          this.getUsers();
        },
        response => {
          this.error = true;
          console.log("ERROR: getCategories: " + JSON.stringify(response.body));
        }
      );
    },

    getUsers: function() {
      var url = "api/v1/users";
      var ids = [];
      for (var i = 0; i < this.categories.length; i++) {
        ids.push(this.categories[i].authorId);
      }
      ids = ids.filter((v, i, a) => a.indexOf(v) === i);
      if (ids.length === 0) {
        this.users = {};
        this.usersReady = true;
        return;
      }
      url += "?ids=" + ids.join(",");
      this.$http.get(url).then(
        response => {
          var users = {};
          for (var i = 0; i < response.body.users.length; i++) {
            users[response.body.users[i].id] = response.body.users[i];
          }
          this.users = users;
          this.usersReady = true;
        },
        response => {
          this.error = true;
          console.log("ERROR: getUsers: " + JSON.stringify(response.body));
        }
      );
    },

    delCategory: function(id) {
      if (!confirm("Are you sure you want to delete category " + id + "?")) {
        return;
      }
      var url = "api/v1/categories/" + id;
      this.$http.delete(url).then(
        response => {
          this.load();
        },
        response => {
          console.log("ERROR: delCategory: " + JSON.stringify(response.body));
        }
      );
    },
  },
});
