<template>
  <section class="manager">
    <header class="manager-head"><div><h3><PosIcon name="search" size="20" /> ISBN Catalog &amp; Editions</h3><p>Search titles, compare editions, and add selected books to the POS cart.</p></div><button class="primary" @click="lookup">Lookup ISBN</button></header>
    <div class="toolbar"><input v-model="query" placeholder="Search ISBN, title, author, or publisher" @keydown.enter="lookup" /><select v-model="edition"><option value="">All editions</option><option>Hardback</option><option>Paperback</option><option>Textbook</option><option>Second-hand</option></select></div>
    <div class="book-grid">
      <article v-for="book in filteredBooks" :key="book.isbn" class="book-card">
        <div class="book-cover"><PosIcon name="file-text" size="28" /></div><div class="book-info"><h4>{{ book.title }}</h4><p>{{ book.author }} · {{ book.publisher }}</p><small>ISBN {{ book.isbn }} · {{ book.edition }}</small><div class="book-footer"><strong>KES {{ book.price.toLocaleString() }}</strong><span :class="book.stock < 5 ? 'warning' : 'ok'">{{ book.stock }} in stock</span></div><button class="secondary" @click="addToCart(book)">Add to cart</button></div>
      </article>
    </div>
    <p v-if="!filteredBooks.length" class="empty">No editions match this search.</p>
  </section>
</template>

<script setup>
import { computed, ref } from 'vue'
import { PosIcon } from '@pos/ui'
const emit = defineEmits(['notification'])
const query = ref('')
const edition = ref('')
const books = ref([
  { isbn: '9789966250124', title: 'The River Between', author: 'Ngũgĩ wa Thiong’o', publisher: 'East African Educational', edition: 'Paperback', price: 850, stock: 18 },
  { isbn: '9780199535569', title: 'Oxford English Dictionary', author: 'Oxford Languages', publisher: 'Oxford University Press', edition: 'Hardback', price: 4200, stock: 4 },
  { isbn: '9789966049117', title: 'Primary Mathematics 6', author: 'Longhorn Kenya', publisher: 'Longhorn', edition: 'Textbook', price: 760, stock: 42 },
  { isbn: '9780141185064', title: 'Animal Farm', author: 'George Orwell', publisher: 'Penguin Classics', edition: 'Second-hand', price: 450, stock: 7 },
])
const filteredBooks = computed(() => books.value.filter((book) => {
  const haystack = `${book.isbn} ${book.title} ${book.author} ${book.publisher}`.toLowerCase()
  return haystack.includes(query.value.toLowerCase()) && (!edition.value || book.edition === edition.value)
}))
function lookup() { emit('notification', { message: query.value ? `Catalog search completed for "${query.value}"` : 'ISBN scanner ready', type: 'info' }) }
function addToCart(book) { window.dispatchEvent(new CustomEvent('pos:add-item', { detail: { id: book.isbn, sku: book.isbn, name: book.title, price_minor: book.price * 100, unit_price_minor: book.price * 100, quantity: 1 } })); emit('notification', { message: `${book.title} added to cart`, type: 'success' }) }
</script>

<style scoped>
.manager { display:flex; flex-direction:column; gap:16px; } .manager-head { display:flex; justify-content:space-between; gap:12px; align-items:flex-start; } h3 { display:flex; align-items:center; gap:8px; margin:0; color:#1e293b; font-size:17px; } p { margin:5px 0 0; color:#64748b; font-size:12px; }.toolbar { display:flex; gap:8px; }.toolbar input,.toolbar select { flex:1; padding:10px; border:1px solid #cbd5e1; border-radius:7px; background:#fff; }.book-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(240px,1fr)); gap:12px; }.book-card { display:flex; gap:12px; padding:13px; border:1px solid #e2e8f0; border-radius:9px; background:#fff; }.book-cover { display:grid; place-items:center; flex:0 0 52px; height:70px; border-radius:6px; background:#ede9fe; color:#7c3aed; }.book-info { flex:1; min-width:0; }.book-info h4 { margin:0; font-size:14px; color:#1e293b; }.book-info p { white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }.book-info small { color:#94a3b8; font-size:10px; }.book-footer { display:flex; justify-content:space-between; margin:9px 0; font-size:12px; }.ok { color:#16a34a; }.warning { color:#d97706; }.primary,.secondary { border:0; border-radius:6px; padding:8px 11px; font-size:11px; font-weight:700; cursor:pointer; }.primary { background:#7c3aed; color:#fff; }.secondary { background:#f3e8ff; color:#6d28d9; }.empty { padding:24px; text-align:center; }
</style>
