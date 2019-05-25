<script>
  import { Router, Link, Route } from "svelte-routing";
  import Home from "./pages/Home.svelte";

  // Auth routes
  import Login from "./pages/Login.svelte";
  import Register from "./pages/Register.svelte";

  // Items routes
  import ReadItem from "./pages/items/Read.svelte";
  import CreateItem from "./pages/items/Create.svelte";
  import Items from "./pages/items/List.svelte";

  export let url = "";
</script>

<!-- App.svelte -->

<Router {url}>
  <nav style="text-align: center;">
    <Link to="/">Home Page</Link>
    <Link to="items">Items</Link>
    <Link to="items/new">Create Item</Link>
    <Link to="webhooks">Webhooks</Link>
    <Link to="login">Login</Link>
    <Link to="register">Register</Link>
  </nav>
  <div>
    <Route path="items" component={Items} />
    <Route path="items/:id" component={ReadItem} />
    <Route path="items/new" component={CreateItem} />
    <Route path="login" component={Login} />
    <Route path="register" component={Register} />
    <Route path="/">
      <Home />
    </Route>
  </div>
</Router>
