<template>
  <div class="mybca-container">
    <div v-if="!isLoggedIn" class="auth-wrapper d-flex align-items-center justify-content-center py-5">
      <div class="card auth-card shadow-lg border-0 rounded-4 p-4 p-md-5 bg-white">
        <div class="text-center mb-4">
          <div class="brand-logo mb-2">my<b>BCA</b> <span class="text-muted fs-6">| Keuangan</span></div>
          <p class="text-muted small">Login untuk akses pencatatan keuangan pribadi</p>
        </div>
        <ul class="nav nav-pills nav-fill mb-4 bg-light p-1 rounded-pill">
          <li class="nav-item"><button @click="authMode = 'login'" class="nav-link rounded-pill py-2 small fw-bold" :class="{ active: authMode === 'login' }">Masuk</button></li>
          <li class="nav-item"><button @click="authMode = 'register'" class="nav-link rounded-pill py-2 small fw-bold" :class="{ active: authMode === 'register' }">Buat Akun</button></li>
        </ul>
        <div v-if="authMode === 'login'">
          <div class="mb-3"><label class="form-label small text-muted">Username</label><input type="text" v-model="authForm.username" class="form-control bg-light shadow-none rounded-3" placeholder="Masukkan username" /></div>
          <div class="mb-4"><label class="form-label small text-muted">Password</label><input type="password" v-model="authForm.password" class="form-control bg-light shadow-none rounded-3" placeholder="Masukkan password" /></div>
          <button @click="handleLogin" class="btn btn-primary w-100 rounded-pill py-2 fw-bold shadow-sm">Masuk</button>
        </div>
        <div v-else>
          <div class="mb-3"><label class="form-label small text-muted">Nama Lengkap</label><input type="text" v-model="authForm.name" class="form-control bg-light shadow-none rounded-3" placeholder="King Ilham" /></div>
          <div class="mb-3"><label class="form-label small text-muted">Username</label><input type="text" v-model="authForm.username" class="form-control bg-light shadow-none rounded-3" placeholder="king_qa" /></div>
          <div class="mb-3"><label class="form-label small text-muted">Password</label><input type="password" v-model="authForm.password" class="form-control bg-light shadow-none rounded-3" placeholder="Min. 6 karakter (Kapital, Angka, Simbol)" /></div>
          <button @click="handleRegister" class="btn btn-success w-100 rounded-pill py-2 fw-bold shadow-sm">Daftar Akun Baru</button>
        </div>
      </div>
    </div>
    <div v-else>
      <header class="mybca-header shadow-sm">
        <div class="container d-flex justify-content-between align-items-center py-3">
          <div class="d-flex align-items-center gap-3">
            <div class="brand-logo">my<b>BCA</b> <span class="badge bg-primary bg-opacity-10 text-primary fs-6 ms-2">Financial Planner</span></div>
          </div>
          <div class="d-flex align-items-center gap-3">
            <span class="user-greeting small">Halo, <b>{{ currentUser.name }}</b></span>
            <button @click="logout" class="btn btn-outline-danger btn-sm rounded-pill px-3">Keluar</button>
          </div>
        </div>
      </header>
      <main class="container my-4">
        <div class="row g-3 mb-4">
          <div class="col-md-4">
            <div class="card border-0 shadow-sm rounded-4 p-4 bg-primary text-white">
              <span class="small text-white-50">Total Pemasukan / Gaji (bulan {{ monthLabel }})</span>
              <h3 class="fw-bold mb-0 mt-1">Rp {{ Number(plan.income || 0).toLocaleString() }}</h3>
            </div>
          </div>
          <div class="col-md-4">
            <div class="card border-0 shadow-sm rounded-4 p-4 bg-white">
              <span class="small text-muted">Total Pengeluaran ({{ filterLabel }})</span>
              <h3 class="fw-bold mb-0 mt-1 text-danger">Rp {{ totalSpent.toLocaleString() }}</h3>
            </div>
          </div>
          <div class="col-md-4">
            <div class="card border-0 shadow-sm rounded-4 p-4 bg-white">
              <span class="small text-muted">Sisa Saldo ({{ filterLabel }})</span>
              <h3 class="fw-bold mb-0 mt-1" :class="totalSisa >= 0 ? 'text-success' : 'text-danger'">Rp {{ totalSisa.toLocaleString() }}</h3>
            </div>
          </div>
        </div>
        <div class="row g-4">
          <div class="col-lg-8">
            <div class="card border-0 shadow-sm rounded-4 p-4 bg-white mb-4">
              <h5 class="fw-bold mb-3 text-dark">📊 Tabel Plotting Keuangan - {{ filterLabel }}</h5>
              <div class="table-responsive">
                <table class="table align-middle mb-0" style="font-size: 14px;">
                  <thead class="table-light">
                    <tr><th>Slot Kategori</th><th>Alokasi %</th><th>Anggaran (Rp)</th><th>Terpakai (Rp)</th><th>Sisa (Rp)</th><th>Status</th></tr>
                  </thead>
                  <tbody>
                    <tr><td><span class="badge bg-primary px-2 py-1">Kebutuhan</span></td><td>{{ plan.pct_kebutuhan }}%</td><td>Rp {{ allocKebutuhan.toLocaleString() }}</td><td>Rp {{ spentKebutuhan.toLocaleString() }}</td><td :class="sisaKebutuhan < 0 ? 'text-danger fw-bold' : 'text-success fw-bold'">Rp {{ sisaKebutuhan.toLocaleString() }}</td><td><span class="badge" :class="sisaKebutuhan >= 0 ? 'bg-success' : 'bg-danger'">{{ sisaKebutuhan >= 0 ? 'Aman' : 'Over' }}</span></td></tr>
                    <tr><td><span class="badge bg-success px-2 py-1">Tabungan</span></td><td>{{ plan.pct_tabungan }}%</td><td>Rp {{ allocTabungan.toLocaleString() }}</td><td>Rp {{ spentTabungan.toLocaleString() }}</td><td :class="sisaTabungan < 0 ? 'text-danger fw-bold' : 'text-success fw-bold'">Rp {{ sisaTabungan.toLocaleString() }}</td><td><span class="badge" :class="sisaTabungan >= 0 ? 'bg-success' : 'bg-danger'">{{ sisaTabungan >= 0 ? 'Aman' : 'Over' }}</span></td></tr>
                    <tr><td><span class="badge bg-warning text-dark px-2 py-1">Keinginan</span></td><td>{{ plan.pct_keinginan }}%</td><td>Rp {{ allocKeinginan.toLocaleString() }}</td><td>Rp {{ spentKeinginan.toLocaleString() }}</td><td :class="sisaKeinginan < 0 ? 'text-danger fw-bold' : 'text-success fw-bold'">Rp {{ sisaKeinginan.toLocaleString() }}</td><td><span class="badge" :class="sisaKeinginan >= 0 ? 'bg-success' : 'bg-danger'">{{ sisaKeinginan >= 0 ? 'Aman' : 'Over' }}</span></td></tr>
                  </tbody>
                </table>
              </div>
              <div v-if="selectedMonth === 0" class="alert alert-info py-2 mt-3 mb-0 small">📅 Menampilkan rekap tahunan {{ selectedYear }} — anggaran dihitung per bulan, total tahunan = anggaran x 12 bisa dilihat di PDF.</div>
            </div>
            <div class="card border-0 shadow-sm rounded-4 p-4 bg-white">
              <div class="d-flex flex-wrap justify-content-between align-items-center mb-3 gap-2">
                <h5 class="fw-bold m-0 text-dark">📋 Riwayat Pengeluaran</h5>
                <div class="d-flex gap-2 align-items-center">
                  <select v-model.number="selectedMonth" class="form-select form-select-sm" style="width: 140px;">
                    <option :value="0">Semua Bulan</option>
                    <option v-for="(m,i) in monthNames" :key="i" :value="i+1">{{ m }}</option>
                  </select>
                  <select v-model.number="selectedYear" class="form-select form-select-sm" style="width: 90px;">
                    <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}</option>
                  </select>
                  <button @click="exportPdf" class="btn btn-sm btn-dark rounded-pill px-3" :disabled="filteredExpenses.length===0">📄 Export PDF</button>
                </div>
              </div>
              <div class="small text-muted mb-2">{{ filterLabel }} — {{ filteredExpenses.length }} transaksi</div>
              <div v-if="filteredExpenses.length === 0" class="text-center py-4 text-muted small">Tidak ada pengeluaran di periode ini.</div>
              <div v-else class="table-responsive" style="max-height: 320px; overflow-y: auto;">
                <table class="table table-sm align-middle mb-0" style="font-size: 13px;">
                  <thead><tr><th>Tanggal</th><th>Kategori</th><th>Keterangan</th><th class="text-end">Nominal</th><th class="text-center">Aksi</th></tr></thead>
                  <tbody>
                    <tr v-for="e in filteredExpenses" :key="e.id"><td>{{ new Date(e.created_at).toLocaleDateString('id-ID') }}</td><td><span class="badge" :class="badgeClass(e.category)">{{ e.category }}</span></td><td>{{ e.description }}</td><td class="text-end fw-bold text-danger">Rp {{ Number(e.amount).toLocaleString('id-ID') }}</td><td class="text-center"><button @click="deleteExpense(e.id)" class="btn btn-sm btn-outline-danger py-0 px-2" style="font-size: 11px;">Hapus</button></td></tr>
                  </tbody>
                </table>
              </div>
              <div v-if="selectedMonth===12 || selectedMonth===0" class="alert alert-warning py-2 mt-3 mb-0 small">💡 Akhir tahun (Desember): gunakan filter <b>Tahun {{ selectedYear }}</b> + <b>Semua Bulan</b> lalu klik Export PDF untuk rekap 12 bulan.</div>
            </div>
          </div>
          <div class="col-lg-4">
            <div class="card border-0 shadow-sm rounded-4 p-4 bg-white mb-4">
              <h5 class="fw-bold mb-3 text-dark">⚙️ Setup Gaji & Slot %</h5>
              <div class="mb-3"><label class="form-label small text-muted">Pemasukan / Gaji (Rp) /bulan</label><div class="input-group"><span class="input-group-text bg-light border-end-0">Rp</span><input type="number" v-model.number="plan.income" class="form-control bg-light border-start-0 shadow-none" placeholder="5000000" /></div></div>
              <div class="row g-2 mb-3">
                <div class="col-4"><label class="form-label small text-muted">Kebutuhan %</label><input type="number" v-model.number="plan.pct_kebutuhan" class="form-control bg-light shadow-none text-center" /></div>
                <div class="col-4"><label class="form-label small text-muted">Tabungan %</label><input type="number" v-model.number="plan.pct_tabungan" class="form-control bg-light shadow-none text-center" /></div>
                <div class="col-4"><label class="form-label small text-muted">Keinginan %</label><input type="number" v-model.number="plan.pct_keinginan" class="form-control bg-light shadow-none text-center" /></div>
              </div>
              <div v-if="pctTotal !== 100" class="alert alert-warning py-1 small mb-3 text-center">Total persen harus 100% (Saat ini: {{ pctTotal }}%)</div>
              <button @click="saveFinancialPlan" class="btn btn-primary w-100 rounded-pill py-2 fw-bold shadow-sm" :disabled="pctTotal !== 100">Simpan Alokasi</button>
            </div>
            <div class="card border-0 shadow-sm rounded-4 p-4 bg-white">
              <h5 class="fw-bold mb-3 text-dark">➕ Catat Pengeluaran</h5>
              <div class="mb-3"><label class="form-label small text-muted">Pilih Kategori Slot</label><select v-model="newExpense.category" class="form-select bg-light shadow-none"><option value="Kebutuhan">Kebutuhan ({{ plan.pct_kebutuhan }}%)</option><option value="Tabungan">Tabungan ({{ plan.pct_tabungan }}%)</option><option value="Keinginan">Keinginan ({{ plan.pct_keinginan }}%)</option></select></div>
              <div class="mb-3"><label class="form-label small text-muted">Nominal (Rp)</label><div class="input-group"><span class="input-group-text bg-light border-end-0">Rp</span><input type="number" v-model.number="newExpense.amount" class="form-control bg-light border-start-0 shadow-none" placeholder="0" /></div></div>
              <div class="mb-3"><label class="form-label small text-muted">Keterangan / Catatan</label><input type="text" v-model="newExpense.description" class="form-control bg-light shadow-none" placeholder="cth: Makan siang, Bensin, Nabung" /></div>
              <button @click="addExpense" class="btn btn-dark w-100 rounded-pill py-2 fw-bold shadow-sm">Simpan Pengeluaran</button>
            </div>
          </div>
        </div>
      </main>
    </div>
    <div v-if="showModal" class="modal-backdrop-custom d-flex align-items-center justify-content-center">
      <div class="modal-card bg-white p-4 rounded-4 shadow-lg text-center animate-pop">
        <h4 class="fw-bold mb-2">{{ modalTitle }}</h4>
        <p class="text-muted small mb-4">{{ modalMessage }}</p>
        <button @click="showModal = false" class="btn btn-dark w-100 rounded-pill py-2 fw-bold">Tutup</button>
      </div>
    </div>
  </div>
</template>
<script>
import axios from 'axios'
import jsPDF from 'jspdf'
const API = 'http://localhost:8080'
export default {
  data() {
    const now = new Date()
    return {
      isLoggedIn: false, authMode: 'login', authForm: { username: '', password: '', name: '' }, currentUser: {},
      showModal: false, modalTitle: '', modalMessage: '',
      plan: { income: 0, pct_kebutuhan: 50, pct_tabungan: 30, pct_keinginan: 20, month: now.getMonth()+1, year: now.getFullYear() },
      expenses: [], newExpense: { category: 'Kebutuhan', amount: 0, description: '' },
      selectedMonth: now.getMonth()+1, selectedYear: now.getFullYear(),
      monthNames: ['Januari','Februari','Maret','April','Mei','Juni','Juli','Agustus','September','Oktober','November','Desember']
    }
  },
  watch: {
    selectedMonth() { if(this.isLoggedIn) this.fetchFinancialPlan() },
    selectedYear() { if(this.isLoggedIn) this.fetchFinancialPlan() }
  },
  computed: {
    pctTotal() { return Number(this.plan.pct_kebutuhan||0) + Number(this.plan.pct_tabungan||0) + Number(this.plan.pct_keinginan||0) },
    allocKebutuhan() { return Math.round(this.plan.income * this.plan.pct_kebutuhan / 100) },
    allocTabungan() { return Math.round(this.plan.income * this.plan.pct_tabungan / 100) },
    allocKeinginan() { return Math.round(this.plan.income * this.plan.pct_keinginan / 100) },
    yearOptions() {
      const y = new Date().getFullYear()
      return [y-1, y, y+1]
    },
    monthLabel() { return this.monthNames[this.selectedMonth-1] || 'Semua' },
    filterLabel() {
      if(this.selectedMonth===0) return `Tahun ${this.selectedYear} (12 bulan)`
      return `${this.monthNames[this.selectedMonth-1]} ${this.selectedYear}`
    },
    filteredExpenses() {
      // backend sudah filter by month/year column — jangan pakai created_at (created_at selalu bulan sekarang)
      return this.expenses.filter(e=>{
        const m = e.month ?? new Date(e.created_at).getMonth()+1
        const y = e.year ?? new Date(e.created_at).getFullYear()
        if(this.selectedMonth===0) return y===this.selectedYear
        return m===this.selectedMonth && y===this.selectedYear
      })
    },
    spentKebutuhan() { return this.filteredExpenses.filter(e => e.category === 'Kebutuhan').reduce((s, e) => s + Number(e.amount), 0) },
    spentTabungan() { return this.filteredExpenses.filter(e => e.category === 'Tabungan').reduce((s, e) => s + Number(e.amount), 0) },
    spentKeinginan() { return this.filteredExpenses.filter(e => e.category === 'Keinginan').reduce((s, e) => s + Number(e.amount), 0) },
    totalSpent() { return this.spentKebutuhan + this.spentTabungan + this.spentKeinginan },
    sisaKebutuhan() { return this.allocKebutuhan - this.spentKebutuhan },
    sisaTabungan() { return this.allocTabungan - this.spentTabungan },
    sisaKeinginan() { return this.allocKeinginan - this.spentKeinginan },
    totalSisa() { return Number(this.plan.income||0) - this.totalSpent }
  },
  methods: {
    openModal(title, message) { this.modalTitle = title; this.modalMessage = message; this.showModal = true; },
    badgeClass(cat) { return cat === 'Kebutuhan' ? 'bg-primary' : cat === 'Tabungan' ? 'bg-success' : 'bg-warning text-dark' },
    async handleLogin() {
      if(!this.authForm.username || !this.authForm.password){ this.openModal('Gagal','Username & Password wajib diisi'); return; }
      try {
        const res = await axios.post(`${API}/api/login`, { username: this.authForm.username, password: this.authForm.password });
        this.currentUser = res.data.user; this.isLoggedIn = true;
        this.plan = { income: 0, pct_kebutuhan: 50, pct_tabungan: 30, pct_keinginan: 20 }; this.expenses = [];
        this.fetchFinancialPlan();
        this.authForm.username=''; this.authForm.password='';
      } catch (err) { this.openModal('Login Gagal', err.response?.data || 'Username atau password salah.'); }
    },
    async handleRegister() {
      if(!this.authForm.username || !this.authForm.password || !this.authForm.name){ this.openModal('Gagal','Semua kolom wajib diisi'); return; }
      try { await axios.post(`${API}/api/register`, { username: this.authForm.username, password: this.authForm.password, name: this.authForm.name }); this.openModal('Berhasil', 'Akun berhasil dibuat. Silakan masuk.'); this.authMode = 'login'; } catch (err) { this.openModal('Registrasi Gagal', err.response?.data || 'Kriteria password tidak terpenuhi.'); }
    },
    async fetchFinancialPlan() {
      try {
        const params = { year: this.selectedYear };
        if (this.selectedMonth !== 0) params.month = this.selectedMonth;
        const res = await axios.get(`${API}/api/financial-plan/${this.currentUser.id}`, { params });
        if (res.data.plan && res.data.plan.user_id !== 0) { this.plan = res.data.plan; } else { this.plan = { income: 0, pct_kebutuhan: 50, pct_tabungan: 30, pct_keinginan: 20, month: this.selectedMonth || new Date().getMonth()+1, year: this.selectedYear }; }
        this.expenses = res.data.expenses || [];
      } catch(e) { this.plan = { income: 0, pct_kebutuhan: 50, pct_tabungan: 30, pct_keinginan: 20, month: this.selectedMonth || new Date().getMonth()+1, year: this.selectedYear }; this.expenses = []; }
    },
    async saveFinancialPlan() {
      if(this.pctTotal !== 100){ this.openModal('Gagal', 'Total persentase harus 100%!'); return; }
      try { await axios.post(`${API}/api/financial-plan`, { user_id: parseInt(this.currentUser.id), month: this.selectedMonth, year: this.selectedYear, income: parseFloat(this.plan.income), pct_kebutuhan: parseFloat(this.plan.pct_kebutuhan), pct_tabungan: parseFloat(this.plan.pct_tabungan), pct_keinginan: parseFloat(this.plan.pct_keinginan) }); this.openModal('Berhasil', 'Alokasi keuangan berhasil disimpan.'); this.fetchFinancialPlan(); } catch(err) { this.openModal('Gagal', err.response?.data || 'Gagal menyimpan alokasi.'); }
    },
    async addExpense() {
      if(!this.newExpense.amount || this.newExpense.amount <= 0){ this.openModal('Gagal', 'Masukkan nominal pengeluaran yang valid.'); return; }
      try { await axios.post(`${API}/api/financial-expense`, { user_id: parseInt(this.currentUser.id), category: this.newExpense.category, amount: parseFloat(this.newExpense.amount), description: this.newExpense.description || this.newExpense.category, month: this.selectedMonth, year: this.selectedYear }); this.newExpense.amount = 0; this.newExpense.description = ''; this.fetchFinancialPlan(); } catch(err) { this.openModal('Gagal', err.response?.data || 'Gagal mencatat pengeluaran.'); }
    },
    async deleteExpense(id) { try { await axios.delete(`${API}/api/financial-expense/${id}`); this.fetchFinancialPlan(); } catch(e) {} },
    exportPdf() {
      const doc = new jsPDF()
      const user = this.currentUser.name || this.currentUser.username || 'User'
      const period = this.filterLabel
      doc.setFontSize(16)
      doc.text(`Rekap Keuangan - ${user}`, 14, 18)
      doc.setFontSize(10)
      doc.text(`Periode: ${period}  |  Dicetak: ${new Date().toLocaleDateString('id-ID')}`, 14, 24)
      doc.setFontSize(11)
      doc.text(`Pemasukan (per bulan): Rp ${Number(this.plan.income).toLocaleString('id-ID')}  |  Alokasi: Kebutuhan ${this.plan.pct_kebutuhan}% / Tabungan ${this.plan.pct_tabungan}% / Keinginan ${this.plan.pct_keinginan}%`, 14, 32)
      // table header
      let y = 40
      doc.setFontSize(9)
      doc.setFillColor(240,240,240)
      doc.rect(14, y, 182, 8, 'F')
      doc.text('Kategori', 16, y+6)
      doc.text('Anggaran', 50, y+6)
      doc.text('Terpakai', 90, y+6)
      doc.text('Sisa', 130, y+6)
      doc.text('Status', 165, y+6)
      y+=8
      const rows = [
        ['Kebutuhan', this.allocKebutuhan, this.spentKebutuhan, this.sisaKebutuhan],
        ['Tabungan', this.allocTabungan, this.spentTabungan, this.sisaTabungan],
        ['Keinginan', this.allocKeinginan, this.spentKeinginan, this.sisaKeinginan],
      ]
      rows.forEach(r=>{
        doc.text(r[0], 16, y+6)
        doc.text(`Rp ${r[1].toLocaleString('id-ID')}`, 50, y+6)
        doc.text(`Rp ${r[2].toLocaleString('id-ID')}`, 90, y+6)
        doc.text(`Rp ${r[3].toLocaleString('id-ID')}`, 130, y+6)
        doc.text(r[3]>=0?'Aman':'Over', 165, y+6)
        y+=8
        doc.line(14, y, 196, y)
      })
      y+=4
      doc.setFontSize(10)
      doc.text(`Total Pengeluaran (${period}): Rp ${this.totalSpent.toLocaleString('id-ID')}  |  Sisa: Rp ${this.totalSisa.toLocaleString('id-ID')}`, 14, y)
      y+=10
      doc.setFontSize(11)
      doc.text(`Riwayat Pengeluaran (${this.filteredExpenses.length} transaksi)`, 14, y)
      y+=6
      doc.setFontSize(8)
      // header
      doc.setFillColor(220,220,220)
      doc.rect(14, y, 182, 7, 'F')
      doc.text('Tanggal', 16, y+5)
      doc.text('Kategori', 38, y+5)
      doc.text('Keterangan', 65, y+5)
      doc.text('Nominal', 160, y+5)
      y+=7
      this.filteredExpenses.forEach(e=>{
        if(y>270){ doc.addPage(); y=14 }
        const d = new Date(e.created_at).toLocaleDateString('id-ID')
        doc.text(d, 16, y+5)
        doc.text(e.category, 38, y+5)
        const desc = (e.description||'').substring(0,38)
        doc.text(desc, 65, y+5)
        doc.text(`Rp ${Number(e.amount).toLocaleString('id-ID')}`, 160, y+5)
        y+=6
      })
      if(this.selectedMonth===0){
        if(y>260){ doc.addPage(); y=14 }
        y+=6
        doc.setFontSize(8)
        doc.text(`* Rekap tahunan ${this.selectedYear}: total pemasukan tahunan estimasi Rp ${(this.plan.income*12).toLocaleString('id-ID')} (Rp ${Number(this.plan.income).toLocaleString('id-ID')} x 12)`, 14, y)
      }
      doc.save(`rekap-keuangan-${user}-${period.replace(/\s+/g,'-')}.pdf`)
    },
    logout() { this.isLoggedIn = false; this.currentUser = {}; this.plan = { income: 0, pct_kebutuhan: 50, pct_tabungan: 30, pct_keinginan: 20 }; this.expenses = []; }
  }
}
</script>
<style scoped>
.mybca-container{background-color:#f8f9fa;min-height:100vh;font-family:'Inter',system-ui,-apple-system,'Segoe UI',Roboto,sans-serif}
.auth-wrapper{min-height:100vh;background:linear-gradient(135deg,#004b93 0%,#0060af 100%)}
.auth-card{width:400px;max-width:90%;border-radius:20px}
.mybca-header{background:#fff;border-bottom:1px solid #eaeaea}
.brand-logo{font-size:1.4rem;color:#0060af;letter-spacing:-0.5px}
.modal-backdrop-custom{position:fixed;top:0;left:0;width:100vw;height:100vh;background:rgba(0,0,0,.4);backdrop-filter:blur(4px);z-index:1050}
.modal-card{width:340px;max-width:90%;border-radius:20px}
@keyframes popIn{0%{transform:scale(.8);opacity:0}100%{transform:scale(1);opacity:1}}
.animate-pop{animation:popIn .25s cubic-bezier(.175,.885,.32,1.275) forwards}
</style>
