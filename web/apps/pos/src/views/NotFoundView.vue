<template>
  <main class="not-found">
    <p class="eyebrow">404</p>
    <h1>Page not found</h1>
    <RouterLink :to="{ name: 'checkout' }">Return to checkout</RouterLink>
  </main>
</template>

<style scoped>
.not-found { max-width:560px; margin:120px auto; padding:24px; text-align:center; }
.not-found h1 { margin:8px 0 16px; }
.not-found a { color:#126b55; font-weight:700; }
</style>
