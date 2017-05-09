const TOPICS_PER_PAGE = 20;

var BebopTopics = Vue.component("bebop-topics", {
  template: `
    <div class="container content-container">

      <div v-if="!dataReady" class="loading-info">
        <div v-if="error" >
          <p class="text-danger">
            Sorry, could not load topics. Please check your connection.
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

        <div class="topics-topic-top-buttons">
          <router-link v-if="auth.authenticated" to="/new-topic" class="btn btn-primary btn-sm">
            <i class="fa fa-plus"></i> New Topic
          </router-link>
          <a class="btn btn-primary btn-sm" role="button" @click="load">
            <i class="fa fa-refresh"></i> Refresh
          </a>
        </div>

        <nav v-if="page > 1">
          <ul class="pagination pagination-sm">
            <li v-for="p in pagination" :class="{active: page === p}">
              <span v-if="p === '...'">…</span>
              <router-link v-if="p !== '...'" :to="'/p/' + p">{{p}}</router-link>
            </li>
          </ul>
        </nav>

        <div v-for="topic in topics" class="card topics-topic">
          <div class="avatar-block">
            <div class="avatar-block-l">
              <img v-if="users[topic.authorId].avatar" class="img-circle" :src="users[topic.authorId].avatar" width="40" height="40"> 
              <img v-else class="img-circle" src="data:image/gif;base64,R0lGODlhAQABAIAAAP///wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw==" width="40" height="40"> 
            </div>
            <div class="avatar-block-r">
              <div class="topics-topic-title">
                <router-link :to="'/t/' + topic.id">{{topic.title}}</router-link>
              </div>
              <div class="topics-topic-info">
                <i class="fa fa-user-o"></i> {{users[topic.authorId].name}}
                <span class="info-separator"> | </span>
                <i class="fa fa-comment-o"></i> {{topic.commentCount}}
                <span class="info-separator"> | </span>
                <i class="fa fa-clock-o"></i> <span :title="topic.lastCommentAt|formatTime">{{topic.lastCommentAt|formatTimeAgo}}</span>
              </div>
              <div class="topics-topic-admin-tools" v-if="auth.authenticated && auth.user.admin">
                <a class="a-tool" role="button" @click="delTopic(topic.id)"><i class="fa fa-times" aria-hidden="true"></i> delete topic</a>
                <span class="info-separator"> | </span> 
                <router-link :to="'/u/' + users[topic.authorId].id" class="a-tool"><i class="fa fa-user" aria-hidden="true"></i> user profile</router-link>
              </div>
            </div>
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
      topics: [],
      topicsReady: false,
      topicCount: 0,
      users: {},
      usersReady: false,
      error: false,
    };
  },

  computed: {
    dataReady: function() {
      return this.topicsReady && this.usersReady;
    },

    page: function() {
      var page = parseInt(this.$route.params.page, 10);
      if (!page || page < 1) {
        return 1;
      }
      return page;
    },

    lastPage: function() {
      if (!this.topicsReady) {
        return 1;
      }
      var p = Math.floor((this.topicCount - 1) / TOPICS_PER_PAGE) + 1;
      if (p < 1) {
        p = 1;
      }
      return p;
    },

    pagination: function() {
      if (!this.topicsReady) {
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
      this.topics = [];
      this.topicsReady = false;
      this.topicCount = 0;
      this.users = {};
      this.usersReady = false;
      this.error = false;
      this.getTopics();
    },

    getTopics: function() {
      var url = "api/v1/topics?limit=" + TOPICS_PER_PAGE;
      if (this.page > 1) {
        var offset = (this.page - 1) * TOPICS_PER_PAGE;
        url += "&offset=" + offset;
      }
      this.$http.get(url).then(
        response => {
          this.topics = response.body.topics;
          this.topicCount = response.body.count;
          this.topicsReady = true;

          if (this.page > this.lastPage) {
            this.$parent.$router.replace("/p/" + this.lastPage);
            return;
          }

          this.getUsers();
        },
        response => {
          this.error = true;
          console.log("ERROR: getTopics: " + JSON.stringify(response.body));
        }
      );
    },

    getUsers: function() {
      var url = "api/v1/users";
      var ids = [];
      for (var i = 0; i < this.topics.length; i++) {
        ids.push(this.topics[i].authorId);
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

    delTopic: function(id) {
      if (!confirm("Are you sure you want to delete topic " + id + "?")) {
        return;
      }
      var url = "api/v1/topics/" + id;
      this.$http.delete(url).then(
        response => {
          this.load();
        },
        response => {
          console.log("ERROR: delTopic: " + JSON.stringify(response.body));
        }
      );
    },
  },
});
