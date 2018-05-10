Vue.component('component-recipe', {
  props:['name', 'idrecipe'],
  template: `
    <div>
      <h1>{{ name }} id {{ idrecipe }}</h1>
      <button v-on:click="'updateRecipeName('+ idrecipe +')'">Update</button>
      <div>
        <a :href="'/recipes/' +idrecipe+ '/image'">
            <img :src="'/recipes/'+ idrecipe +'/image'"/>
        </a>
      </div>      
    </div>
    `
})

var app = new Vue({
  
  el: '#showRecipe',
 
  created: function() {
    console.log("new recipe created");
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
