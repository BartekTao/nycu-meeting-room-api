<template>
  <background-header></background-header>

  <router-view></router-view>
</template>

<style lang="scss">
@import "bootstrap";
</style>

<style scoped>
@import '@/assets/css/fontawesome.css';
@import '@/assets/css/templatemo-plot-listing.css';
@import '@/assets/css/animated.css';
@import '@/assets/css/customization.css';
@import '@/assets/css/bootstrap.min.css';
</style>

<script>

import BackgroundHeader from '@/components/BackgroundHeader.vue';

export default {
  name: 'App',
  data () {
    return {
      first_data: '',
    };
  },
  components: {
    BackgroundHeader,
  },
}


</script>