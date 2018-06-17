Vue.component("component-recipe", {
  data: function() {
    return {
      recipeName: this.initialRecipeName,
      recipeId: this.initialRecipeId,
      show: false
    };
  },

  props: ["initialRecipeName", "initialRecipeId"],
  template: `
    <div>
      <div v-show="!show">
        <h1>{{ recipeName }}</h1>
        <button v-on:click="show = !show" >Edit</button>
      </div>
      <div v-show="show">
        <input v-model="recipeName" />      
        <button v-on:click="updateRecipeName()">Update</button>
      </div>
      <div>
        <a :href="'/recipes/' +recipeId+ '/image'">
            <img :src="'/recipes/'+ recipeId +'/image'"/>
        </a>
      </div>      
    </div>
    `,

  methods: {
    updateRecipeName: function() {
      console.log("recipe id:" + this.recipeId + "name: " + this.recipeName);
      // POST /someUrl
      this.$http
        .post("/recipes/" + this.recipeId, {
          ID: this.recipeId,
          RecipeName: this.recipeName
        })
        .then(
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

var app = new Vue({
  el: "#showRecipe",

  created: function() {
    console.log("new recipe created");
  }
});
