var ImageUploader = VueImageUploadResize.ImageUploader;

var app = new Vue({
  el: "#newRecipe",

  data: {
    name: "recipe name 1",
    hasImage: false,
    image: {}
  },

  methods: {
    setImage: function(file) {
      this.hasImage = true;
      this.image = file;
      console.log("file:" + file);
      formData = new FormData();
      formData.append("name", this.name);
      formData.append("uploadfile", this.image);
      this.$http.post("/recipes", formData).then(
        response => {
          this.show = false;
        },
        response => {
          // error callback
        }
      );
    }
  }
});
