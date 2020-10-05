import { RouteConfig } from 'vue-router/types/router';
import Layout from '@/layout/index.vue';

export const enum backendRoutes {
  USER_REGISTRATION = "/users/",
  USER_AUTH_STATUS = "/users/status",
  LOGIN = "/users/login",
  LOGOUT = "/users/logout",
  VERIFY_2FA_SECRET = "/users/totp_secret/verify",
  VALID_INGREDIENTS = "/api/v1/valid_ingredients",
  VALID_INGREDIENT = "/api/v1/valid_ingredients/{}",
  VALID_INSTRUMENTS = "/api/v1/valid_instruments",
  VALID_INSTRUMENT = "/api/v1/valid_instruments/{}",
  VALID_PREPARATIONS = "/api/v1/valid_preparations",
  VALID_PREPARATION = "/api/v1/valid_preparations/{}",
}

const catchAllRoute: RouteConfig = {
  path: "*",
  redirect: "/404",
  meta: {
    title: '',
    hidden: true,
  },
};

const landingRoute: RouteConfig = {
  path: "/",
  name: "landing",
  component: () => import(/* webpackChunkName: "register" */ "@/views/landing/index.vue"),
  meta: {
    title: 'Landing page',
    hidden: true,
  },
};

const registrationRoute: RouteConfig = {
  path: "/register",
  name: "register",
  component: () => import(/* webpackChunkName: "register" */ "@/views/register/index.vue"),
  meta: {
    title: 'Register a new account',
    hidden: true,
  },
};

const loginRoute: RouteConfig = {
  path: "/login",
  name: "login",
  component: () => import(/* webpackChunkName: "login" */ "@/views/login/index.vue"),
  meta: {
    title: 'Login',
    hidden: true,
  },
};

const logoutRoute: RouteConfig = {
  path: "/logout",
  name: "logout",
  component: () => import(/* webpackChunkName: "logout" */ "@/views/logout/index.vue"),
  meta: {
    title: 'Logout',
    hidden: true,
  },
};

const notFoundRoute: RouteConfig = {
  path: "/404",
  name: "notFound",
  component: () => import(/* webpackChunkName: "404" */ "@/views/404.vue"),
  meta: {
    title: 'Not Found',
    hidden: true,
  },
};

const adminDashboardRoute: RouteConfig =  {
    path: "/admin/dashboard",
    component: () => import(/* webpackChunkName: "dashboard" */ "@/views/dashboard/index.vue"),
    meta: {
      title: "Dashboard",
      icon: "dashboard",
      name: "dashboard",
    },
};

const homeRoute: RouteConfig = {
  path: "/admin",
  component: Layout,
  // beforeEnter: (to: Route, from: Route, next: NavigationGuardNext) => {
  //   console.log(to, from);
  // },
  name: "home",
  redirect: "/admin/dashboard",
  children: [adminDashboardRoute],
};

// plural enumerations
const validIngredientTagsRoute: RouteConfig =  {
  name: "validIngredientTags",
  path: "/admin/enumerations/valid_ingredient_tags",
  component: () => import(/* webpackChunkName: "validIngredients" */ "@/views/validIngredients/validIngredients.vue"),
  meta: {
    title: "Ingredient Tags",
    icon: "table",
  },
};

const validIngredientPreparationsRoute: RouteConfig =  {
  name: "validIngredientPreparations",
  path: "/admin/enumerations/valid_ingredient_preparations",
  component: () => import(/* webpackChunkName: "validIngredientPreparations" */ "@/views/validIngredients/validIngredients.vue"),
  meta: {
    title: "Ingredient Preparations",
    icon: "table",
  },
};

//// individual instance routes for enumerable objects

// Ingredients

const validIngredientsRoute: RouteConfig =  {
  name: "validIngredients",
  path: "/admin/enumerations/valid_ingredients",
  component: () => import(/* webpackChunkName: "validIngredients" */ "@/views/validIngredients/validIngredients.vue"),
  meta: {
    title: "Ingredients",
    icon: "table",
  },
};

const validIngredientRoute: RouteConfig = {
  name: "validIngredient",
  path: "/admin/enumerations/valid_ingredients/:validIngredientID",
  component: () => import(/* webpackChunkName: "validIngredientTable" */ "@/views/validIngredients/validIngredient.vue"),
  meta: {
    hidden: true,
  },
};

const createValidIngredientRoute: RouteConfig = {
  name: "createValidIngredient",
  path: "/admin/enumerations/valid_ingredients/new",
  component: () => import(/* webpackChunkName: "createValidIngredient" */ "@/views/validIngredients/create.vue"),
  meta: {
    hidden: true,
  },
};

// Instruments

const validInstrumentsRoute: RouteConfig =  {
  name: "validInstruments",
  path: "/admin/enumerations/valid_instruments",
  component: () => import(/* webpackChunkName: "validInstruments" */ "@/views/validInstruments/validInstruments.vue"),
  meta: {
    title: "Instruments",
    icon: "table",
  },
};

const validInstrumentRoute: RouteConfig = {
  name: "validInstrument",
  path: "/admin/enumerations/valid_instruments/:validInstrumentID",
  component: () => import(/* webpackChunkName: "validInstrument" */ "@/views/validInstruments/validInstrument.vue"),
  meta: {
    hidden: true,
  },
};

const createValidInstrumentRoute: RouteConfig = {
  name: "createValidInstrument",
  path: "/admin/enumerations/valid_instruments/new",
  component: () => import(/* webpackChunkName: "createValidInstrument" */ "@/views/validInstruments/create.vue"),
  meta: {
    hidden: true,
  },
};

// Preparations

const validPreparationsRoute: RouteConfig =  {
  name: "validPreparations",
  path: "/admin/enumerations/valid_preparations",
  component: () => import(/* webpackChunkName: "validPreparationsTable" */ '@/views/validPreparations/validPreparations.vue'),
  meta: {
    title: "Preparations",
    icon: "table",
  },
};

const validPreparationRoute: RouteConfig = {
  name: "validPreparation",
  path: "/admin/enumerations/valid_preparations/:validPreparationID",
  component: () => import(/* webpackChunkName: "validPreparation" */ '@/views/validPreparations/validPreparation.vue'),
  meta: {
    hidden: true,
  },
};

const createValidPreparationRoute: RouteConfig = {
  name: "createValidPreparation",
  path: "/admin/enumerations/valid_preparations/new",
  component: () => import(/* webpackChunkName: "createValidPreparation" */ '@/views/validPreparations/create.vue'),
  meta: {
    hidden: true,
  },
};


const enumerations: RouteConfig = {
  name: "enumerations",
  path: "/admin/enumerations",
  component: Layout,
  meta: {
    title: "Enumerations",
    icon: "example",
  },
  children: [
    createValidIngredientRoute,
    validIngredientsRoute,
    validIngredientRoute,
    createValidInstrumentRoute,
    validInstrumentsRoute,
    validInstrumentRoute,
    createValidPreparationRoute,
    validPreparationsRoute,
    validPreparationRoute,
    validIngredientTagsRoute,
    validIngredientPreparationsRoute,
  ],
};

const createRecipeStepExperimentRoute: RouteConfig = {
  name: "createRecipeStep",
  path: "/admin/experiments/recipe_steps/new",
  component: () => import(/* webpackChunkName: "createRecipe" */ '@/views/recipeSteps/create.vue'),
  meta: {
    title: "Create Recipe Step Experiment",
    icon: "dashboard",
    name: "dashboard",
  },
};

const createRecipeExperimentRoute: RouteConfig = {
  name: "createRecipe",
  path: "/admin/experiments/recipes/new",
  component: () => import(/* webpackChunkName: "createRecipe" */ '@/views/recipes/create.vue'),
  meta: {
    title: "Create Recipe Experiment",
    icon: "dashboard",
    name: "dashboard",
  },
};

export const routes: RouteConfig[] = [
  homeRoute,
  registrationRoute,
  loginRoute,
  logoutRoute,
  notFoundRoute,
  adminDashboardRoute,
  // experiments
  createRecipeStepExperimentRoute,
  createRecipeExperimentRoute,
  // big boys and girls
  enumerations,
  landingRoute,
  catchAllRoute,
  // adminRoute
];


const sidebarEnumerations: RouteConfig = {
  name: "enumerations",
  path: "/admin/enumerations",
  component: Layout,
  meta: {
    title: "Enumerations",
    icon: "example",
  },
  children: [
    validIngredientsRoute,
    validInstrumentsRoute,
    validPreparationsRoute,
  ],
};

export const sidebarRoutes: RouteConfig[] = [
  adminDashboardRoute,
  sidebarEnumerations,
  createRecipeExperimentRoute,
];
