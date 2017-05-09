Vue.component("bebop-nav", {
  template: `
    <nav class="navbar navbar-default navbar-fixed-top">
      <div class="container">
        <div class="navbar-header pull-left">
          <router-link to="/" class="navbar-brand">
            <span class="navbar-title">
              <i class="fa fa-comments"></i>
              {{ config.title }}
            </span>
          </router-link>
        </div>
        <div class="navbar-header pull-right">
          <ul class="nav pull-left">
            <li v-if="auth.authenticated">
              <a class="navbar-link dropdown-toggle navbar-user" role="button" data-toggle="dropdown" :title="auth.user.name">
                <img v-if="auth.user.avatar" class="img-circle" :src="auth.user.avatar" width="35" height="35"> 
                <img v-else class="img-circle" src="data:image/gif;base64,R0lGODlhAQABAIAAAP///wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw==" width="35" height="35"> 
                <span class="caret"></span>
              </a>
              <ul class="dropdown-menu pull-right">
                <li>
                  <router-link to="/me">
                    <i class="fa fa-user icon-s"></i>
                    {{auth.user.name}}
                  </router-link>
                </li>
                <li role="separator" class="divider"></li>
                <li>
                  <a href="#" @click.prevent="$parent.signOut()">
                    <i class="fa fa-sign-out icon-s"></i>
                    Sign out
                  </a>
                </li>
              </ul>
            </li>
            <li v-else>
              <a class="navbar-link dropdown-toggle navbar-sign-in" href="#" data-toggle="dropdown">
                <i class="fa fa-user icon-s"></i>
                Sign In / Up 
                <span class="caret"></span>
              </a>
              <ul class="dropdown-menu pull-right">
                <li v-for="provider in config.oauth">
                  <a href="#" @click.prevent="$parent.signIn(provider)">
                    <i :class="'icon-s fa fa-' + provider" aria-hidden="true"></i>
                    with {{provider|capitalize}}
                  </a>
                </li>
              </ul>
            </li>
          </ul>
        </div>
      </div>
    </nav>
  `,

  props: ["config", "auth"],

  data: function() {
    return {};
  },
});
