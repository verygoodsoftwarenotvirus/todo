<script lang="typescript">
  // make dynamic date to be added to footer
  let date: Number = new Date().getFullYear();
</script>

<footer class="block py-4">
  <div class="container mx-auto px-4">
    <hr class="mb-4 border-b-1 border-gray-300" />
    <div class="flex flex-wrap items-center md:justify-between justify-center">
      <div class="w-full md:w-4/12 px-4">
        <div
          class="text-sm text-gray-600 font-semibold py-1 text-center md:text-left"
        >
          Copyright © {date}
        </div>
      </div>
      <div class="w-full md:w-8/12 px-4">
        <ul class="flex flex-wrap list-none md:justify-end justify-center">
          <li>
            <a
              href="#"
              class="text-gray-700 hover:text-gray-900 text-sm font-semibold block py-1 px-3"
            >
              About Us
            </a>
          </li>
          <li>
            <a
              href="#"
              class="text-gray-700 hover:text-gray-900 text-sm font-semibold block py-1 px-3"
            >
              Blog
            </a>
          </li>
        </ul>
      </div>
    </div>
  </div>
</footer>
