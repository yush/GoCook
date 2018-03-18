var app = new Vue({
  
  el: '#showRecipe',
 
  mounted: function() {
    console.log("new recipe mounted");
  },
  
  data: {
    recipeName: 'baba',
    message: 'This string was updated on: ' + new Date()
  },

  methods: {
    updateRecipeName: function(aRecipeID) {
      console.log("recipe id:"+ aRecipeID+ "name: "+this.$data.recipeName);

      // POST /someUrl
      this.$http.post('/recipes/'+aRecipeID, {ID: aRecipeID, RecipeName: this.$data.recipeName}).then(response => {

        // get status
        response.status;

        // get status text
        response.statusText;

        // get 'Expires' header
        response.headers.get('Expires');

        // get body data
        this.someData = response.body;

      }, response => {
        // error callback
      });
    },
  }
})
