<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1600px] space-y-5">
      <header class="flex flex-wrap items-center justify-between gap-3">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('tests.title') }}</h2>
        <button class="btn btn-secondary" :disabled="loading" @click="load(true)"><Icon name="refresh" size="sm" />{{ t('common.refresh') }}</button>
      </header>

      <p v-if="auth.isAdmin && reviewError" role="alert" class="text-sm text-red-600 dark:text-red-400">{{ reviewError }}</p>

      <div v-if="displayResults.length" class="flex flex-wrap items-end justify-between gap-3 border-b border-gray-200 dark:border-dark-700">
        <nav class="flex min-w-0 flex-1 gap-1 overflow-x-auto" role="tablist" :aria-label="t('tests.groupFilter')">
          <button v-for="group in availableGroups" :id="tabId(group.key)" :key="group.key" type="button" role="tab" :aria-selected="activeGroup === group.key" aria-controls="quality-results" :tabindex="activeGroup === group.key ? 0 : -1" class="shrink-0 border-b-2 px-4 py-3 text-base font-semibold transition-colors" :class="activeGroup === group.key ? 'border-primary-500 text-primary-600 dark:text-primary-400' : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200'" @click="activeGroup = group.key" @keydown="onGroupKeydown($event, group.key)">{{ group.name }}</button>
        </nav>
        <label class="mb-2 flex w-full items-center gap-2 text-sm text-gray-500 dark:text-gray-400 sm:w-auto">
          <span class="shrink-0">{{ t('tests.modelFilter') }}</span>
          <select v-model="modelFilter" class="input min-w-0 flex-1 text-sm sm:w-48"><option value="">{{ t('tests.allModels') }}</option><option v-for="model in availableModels" :key="model" :value="model">{{ model }}</option></select>
        </label>
      </div>

      <div v-if="votesLoadError" class="rounded-md border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800 dark:border-amber-900 dark:bg-amber-950/20 dark:text-amber-200" role="alert" data-votes-error>{{ votesLoadError }}<button type="button" class="ml-3 underline" :disabled="loading" @click="load(true)">{{ t('common.retry') }}</button></div>
      <p v-if="!displayResults.length" class="py-16 text-center text-sm text-gray-500">{{ loading ? t('common.loading') : t('tests.empty') }}</p>
      <section v-else id="quality-results" ref="resultsPanel" class="min-w-0 space-y-5" role="tabpanel" :aria-labelledby="tabId(activeGroup)" tabindex="0">
        <section v-if="standaloneVotes.length" class="space-y-4" data-voting-section>
          <div><h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('tests.voting.title') }}</h3><p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('tests.voting.description') }}</p></div>
          <TestVoteCard v-for="item in standaloneVotes" :key="item.result.id" :item="item" :busy="votingResultIds.has(item.result.id)" :error="voteErrors[item.result.id]" @vote="castVote(item.result.id, $event)" />
        </section>
        <p v-if="!accountResults.length && !standaloneVotes.length" class="py-16 text-center text-sm text-gray-500">{{ t('tests.noMatches') }}</p>
        <div v-else-if="accountResults.length" class="grid min-w-0 items-start gap-4 lg:grid-cols-[18rem_minmax(0,1fr)] xl:grid-cols-[20rem_minmax(0,1fr)]" data-quality-workspace>
          <aside class="min-w-0 overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900 lg:sticky lg:top-4" data-history-sidebar>
            <div class="border-b border-gray-200 px-4 py-3 dark:border-dark-700">
              <div class="flex items-center justify-between gap-3">
                <div><h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('tests.historyList') }}</h3><p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('tests.historyListHint') }}</p></div>
                <span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs tabular-nums text-gray-600 dark:bg-dark-800 dark:text-gray-300">{{ accountResults.reduce((count, account) => count + account.tests.length, 0) }}</span>
              </div>
            </div>
            <div class="max-h-64 overflow-y-auto p-2 lg:max-h-[32vh]" data-series-list>
              <article v-for="(account, accountIndex) in accountResults" :key="account.key" class="mb-2 overflow-hidden rounded-lg border border-gray-100 dark:border-dark-800" data-account-result>
                <header class="bg-gray-50/70 px-3 py-2 dark:bg-dark-800/50" data-account-info><h3 class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ t('tests.resultSet', { number: accountIndex + 1 }) }}</h3><p class="truncate text-xs text-gray-500 dark:text-gray-400">{{ account.groupName }}</p></header>
                <div class="space-y-1 p-2" data-account-tests>
                  <section
                    v-for="test in account.tests"
                    :key="test.key"
                    class="min-w-0 rounded-md border border-transparent p-2 transition-colors hover:border-gray-200 hover:bg-gray-50 dark:hover:border-dark-700 dark:hover:bg-dark-800/60"
                    :class="selectedSeriesId === seriesIdentityFor(account, test) ? 'border-primary-200 bg-primary-50/70 dark:border-primary-900 dark:bg-primary-950/20' : ''"
                    data-test-section
                    :data-numeric-test="test.latest.output_kind === 'number' ? '' : undefined"
                    :data-statistics-test="test.latest.output_kind === 'statistics' ? '' : undefined"
                    :data-content-test="!['number', 'statistics'].includes(test.latest.output_kind) ? '' : undefined"
                  >
                    <div class="flex min-w-0 items-start gap-2">
                      <button type="button" class="min-w-0 flex-1 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500" :aria-pressed="selectedSeriesId === seriesIdentityFor(account, test)" @click="selectSeries(account, test)">
                        <h4 class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ test.name }}</h4>
                        <div class="mt-1 flex min-w-0 flex-wrap items-center gap-1.5"><span :class="statusClass(test.latest)">{{ statusLabel(test.latest) }}</span><span v-if="voteFor(test.latest)" class="rounded-full bg-amber-50 px-1.5 py-0.5 text-[10px] font-medium text-amber-700 dark:bg-amber-950/30 dark:text-amber-300" data-vote-entry>{{ t('tests.voting.title') }}</span><span class="min-w-0 truncate text-xs text-gray-500 dark:text-gray-400">{{ modelLabel(test.latest) }}<span v-if="test.latest.output_kind !== 'statistics' && test.latest.latency_ms != null"> · {{ test.latest.latency_ms }}ms</span></span></div>
                        <p class="mt-1 text-[11px] text-gray-400 dark:text-gray-500">{{ formatDate(resultTime(test.latest)) }}</p>
                        <strong v-if="test.latest.output_kind === 'number'" class="mt-2 block break-all text-xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ test.latest.output_numeric ?? '-' }}</strong>
                      </button>
                    </div>
                  </section>
                </div>
              </article>
            </div>
            <div v-if="selectedSeries" class="border-t border-gray-200 dark:border-dark-700" data-history-records>
              <div class="flex items-center justify-between gap-2 px-4 py-3"><div class="min-w-0"><h4 class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ selectedSeries.test.name }}</h4><p class="truncate text-xs text-gray-500 dark:text-gray-400">{{ t('tests.historyRecords') }}</p></div><span v-if="historyLoading" role="status" :aria-label="t('common.loading')" class="h-3 w-3 animate-spin rounded-full border-2 border-primary-500 border-t-transparent" /></div>
              <div class="max-h-64 space-y-1 overflow-y-auto px-2 pb-2">
                <button v-for="result in workspaceHistoryItems" :key="result.id" type="button" class="w-full rounded-md border border-transparent px-3 py-2 text-left transition-colors hover:bg-gray-50 dark:hover:bg-dark-800" :class="selectedResultId === result.id ? 'border-primary-200 bg-primary-50 dark:border-primary-900 dark:bg-primary-950/30' : ''" :aria-current="selectedResultId === result.id ? 'true' : undefined" :data-history-record="result.id" @click="selectResult(result)">
                  <span class="flex items-center justify-between gap-2"><span :class="statusClass(result)">{{ statusLabel(result) }}</span><span class="text-[11px] text-gray-500 dark:text-gray-400">{{ formatDate(resultTime(result)) }}</span></span>
                  <span class="mt-1 block truncate text-xs text-gray-500 dark:text-gray-400">{{ modelLabel(result) }}<span v-if="result.output_kind !== 'statistics' && result.latency_ms != null"> · {{ result.latency_ms }}ms</span></span>
                </button>
              </div>
              <div v-if="historyError" class="flex flex-wrap items-center gap-2 px-4 pb-3 text-xs"><span class="text-red-600 dark:text-red-400" role="alert">{{ historyError }}</span><button type="button" class="font-medium text-primary-600 hover:underline" data-history-retry :disabled="historyLoading" @click="loadHistory()">{{ t('common.retry') }}</button></div>
              <div v-else-if="historyBeforeId" class="flex justify-center px-3 pb-3"><button type="button" class="btn btn-secondary btn-sm" data-history-more :disabled="historyLoading" @click="loadHistory()">{{ t('tests.loadMore') }}</button></div>
            </div>
          </aside>

          <section class="min-h-0 min-w-0 overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900" data-result-detail>
            <template v-if="selectedSeries">
              <header class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-gray-200 px-5 py-4 dark:border-dark-700 sm:px-6" data-result-detail-header>
                <div class="min-w-0"><h3 class="truncate text-lg font-semibold text-gray-900 dark:text-white">{{ selectedSeries.test.name }}</h3><p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('tests.historyRecords') }}</p></div>
                <span class="shrink-0 rounded-full bg-gray-100 px-2.5 py-1 text-xs tabular-nums text-gray-600 dark:bg-dark-800 dark:text-gray-300">{{ workspaceHistoryItems.length }}</span>
              </header>
              <div ref="resultGallery" class="max-h-[calc(100vh-12rem)] overflow-y-auto p-4 sm:p-5" data-result-gallery>
                <div class="grid min-w-0 grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4" data-result-grid>
                  <article v-for="result in workspaceHistoryItems" :key="result.id" class="min-w-0 overflow-hidden rounded-lg border bg-white transition-shadow dark:bg-dark-900" :class="selectedResultId === result.id ? 'border-primary-400 shadow-md dark:border-primary-500' : 'border-gray-200 dark:border-dark-700'" :data-result-card="result.id">
                    <header class="border-b border-gray-100 px-3 py-3 dark:border-dark-800">
                      <div class="flex min-w-0 items-center justify-between gap-2"><span :class="statusClass(result)">{{ statusLabel(result) }}</span><span class="shrink-0 text-[11px] text-gray-500 dark:text-gray-400">{{ formatDate(resultTime(result)) }}</span></div>
                      <p class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">{{ modelLabel(result) }}<span v-if="result.output_kind !== 'statistics' && result.latency_ms != null"> · {{ result.latency_ms }}ms</span></p>
                    </header>
                    <figure class="min-w-0 p-3">
                      <template v-if="result.output_kind === 'number'"><div class="flex min-h-28 items-center justify-center rounded-lg bg-gray-50 p-4 dark:bg-dark-800" data-numeric-gallery><strong class="break-all text-3xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ result.output_numeric ?? '-' }}</strong></div></template>
                      <TestResultOutput v-else :result="result" compact :max-preview-height="260" />
                      <figcaption v-if="result.output_kind === 'html' && result.output_html" class="mt-3 flex justify-end text-xs"><button type="button" class="font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400" data-expand-preview @click="previewTarget = result">{{ t('tests.expandPreview') }}</button></figcaption>
                      <AdminTestDecision v-if="reviewFor(result)" :review="reviewFor(result)!" compact @decided="refreshAfterDecision" />
                    </figure>
                    <section v-if="voteFor(result)" class="border-t border-gray-100 p-3 dark:border-dark-800" data-voting-section>
                      <TestVoteCard :item="voteFor(result)!" :busy="votingResultIds.has(result.id)" :error="voteErrors[result.id]" :show-output="false" embedded @vote="castVote(result.id, $event)" />
                    </section>
                  </article>
                </div>
                <p v-if="historyLoading" class="py-6 text-center text-sm text-gray-500" role="status">{{ t('common.loading') }}</p>
                <div v-else-if="historyError" class="flex flex-wrap items-center justify-center gap-3 py-6 text-sm"><p class="text-red-600 dark:text-red-400" role="alert">{{ historyError }}</p><button type="button" class="btn btn-secondary btn-sm" data-history-retry :disabled="historyLoading" @click="loadHistory()">{{ t('common.retry') }}</button></div>
                <div v-else-if="historyBeforeId" class="flex justify-center pt-5"><button type="button" class="btn btn-secondary btn-sm" data-history-more :disabled="historyLoading" @click="loadHistory()">{{ t('tests.loadMore') }}</button></div>
              </div>
            </template>
            <p v-else class="p-10 text-center text-sm text-gray-500">{{ t('tests.noMatches') }}</p>
          </section>
        </div>
      </section>
    </div>

    <BaseDialog :show="!!previewTarget" :title="previewTarget ? `${groupName(previewTarget)} · ${testName(previewTarget)}` : ''" width="extra-wide" @close="previewTarget = null">
      <template v-if="previewTarget">
        <TestResultOutput :result="previewTarget" />
        <AdminTestDecision v-if="reviewFor(previewTarget)" :review="reviewFor(previewTarget)!" @decided="refreshAfterDecision" />
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { testResultsAPI } from '@/api/testResults'
import { listReviews, listResults as listAdminResults, type TestAdminReview } from '@/api/admin/tests'
import type { TestResult, TestVote, TestVoteResult } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import TestResultOutput from '@/components/tests/TestResultOutput.vue'
import AdminTestDecision from '@/components/tests/AdminTestDecision.vue'
import TestVoteCard from '@/components/tests/TestVoteCard.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { extractApiErrorMessage } from '@/utils/apiError'

interface TestSeries {
  key: string
  name: string
  order: number
  latest: TestResult
  results: TestResult[]
}

interface TestAccountSeries {
  key: string
  name: string
  groupName: string
  order: number
  tests: TestSeries[]
  schedulable?: boolean
  rank: number
}

const { t } = useI18n()
const route = useRoute()
const app = useAppStore()
const auth = useAuthStore()
const loading = ref(false)
const allResults = ref<TestResult[]>([])
const voteResults = ref<TestVoteResult[]>([])
const votesLoadError = ref('')
const votingResultIds = ref(new Set<number>())
const voteErrors = ref<Record<number, string>>({})
let dataRevision = 0
let unmounted = false
let loadPromise: Promise<void> | undefined
const adminReviews = ref<TestAdminReview[]>([])
const reviewError = ref('')
const reviewIndex = computed(() => new Map(auth.isAdmin ? adminReviews.value.map(review => [review.result.id, review]) : []))
const reviewFor = (result: TestResult) => reviewIndex.value.get(result.id)
// Reviews and votes provide a fallback when a result has no live scheduling
// projection. A paused account remains visible in the regular result list.
const pausedAccountIds = computed(() => {
  const ids = new Set<number>()
  for (const review of adminReviews.value) {
    if (review.account_paused && review.result.account_id != null) ids.add(review.result.account_id)
  }
  for (const item of voteResults.value) {
    if (item.voting.account_paused && item.result.account_id != null) ids.add(item.result.account_id)
  }
  return ids
})
const isHiddenPausedAccount = (result: TestResult) => result.account_id != null && pausedAccountIds.value.has(result.account_id)
const visibleVoteResults = computed(() => voteResults.value.filter(item => !isHiddenPausedAccount(item.result)
  && (!requestedGroupID.value || item.result.group_id === requestedGroupID.value)))
const regularResults = computed(() => {
  const records = new Map(allResults.value.map(result => [result.id, result]))
  // The normal result endpoint contains the freshest group projection. Active
  // review records are merged only when that result is unavailable; an older
  // review payload must not move a freshly re-routed account back to its
  // source group until the next poll.
  if (auth.isAdmin) for (const review of adminReviews.value.filter(item => !requestedGroupID.value || item.result.group_id === requestedGroupID.value)) {
    const current = records.get(review.result.id)
    records.set(review.result.id, current ? { ...review.result, ...current, account_name: review.result.account_name } : review.result)
  }
  return [...records.values()]
})
const displayResults = computed(() => {
  const records = new Map(regularResults.value.map(result => [result.id, result]))
  for (const item of visibleVoteResults.value) if (!records.has(item.result.id)) records.set(item.result.id, item.result)
  return [...records.values()]
})
const activeGroup = ref('')
const modelFilter = ref('')
const selectedSeriesId = ref('')
const selectedResultId = ref<number>()
const followLatestSelection = ref(true)
const previewTarget = ref<TestResult | null>(null)
const historyResults = ref<TestResult[]>([])
const historyLoading = ref(false)
const historyError = ref('')
const historyBeforeId = ref<number>()
const historyLoadedSeriesId = ref('')
let adminHistoryLimit = 20
let historyRequest = 0
const resultsPanel = ref<HTMLElement | null>(null)
const resultGallery = ref<HTMLElement | null>(null)
watch([activeGroup, modelFilter], () => {
  if (resultsPanel.value) resultsPanel.value.scrollTop = 0
}, { flush: 'post' })

const resultTime = (result: TestResult) => result.started_at || result.finished_at || result.created_at
const resultTimestamp = (result: TestResult) => new Date(resultTime(result) || 0).getTime()
const sortResults = (items: TestResult[]) => [...items].sort((a, b) => resultTimestamp(b) - resultTimestamp(a) || b.id - a.id)
const testName = (result: TestResult) => result.test_name || result.plan_name || result.test_definition?.name || result.plan?.name || t('tests.unknownType')
const orderValue = (value?: number) => typeof value === 'number' && Number.isFinite(value) ? value : Number.MAX_SAFE_INTEGER
const exposesAccount = (result: TestResult) => result.target_mode !== 'group' && (result.result_set_id != null || result.account_id != null)
const targetKey = (result: TestResult) => requestedGroupID.value > 0 ? 'group' : exposesAccount(result) ? `account:${result.result_set_id ?? result.account_id}` : 'group'
const groupKey = (result: TestResult) => result.group_id != null ? `group:${result.group_id}` : result.group_name ? `name:${result.group_name}` : 'ungrouped'
const groupName = (result: TestResult) => result.group_name || (result.group_id != null ? `${t('tests.group')} #${result.group_id}` : t('tests.ungrouped'))
const testKey = (result: TestResult) => `${result.test_definition_id ?? result.test_definition?.id ?? testName(result)}:${result.model_id || ''}:${result.reasoning_effort || ''}`
const tabId = (key: string) => `quality-group-${encodeURIComponent(key)}`
const requestedGroupID = computed(() => {
  const raw = Array.isArray(route.query.group_id) ? route.query.group_id[0] : route.query.group_id
  return raw && /^\d+$/.test(raw) ? Number(raw) : 0
})

const availableGroups = computed(() => {
  const groups = new Map<string, { key: string; name: string; order: number }>()
  for (const result of displayResults.value) {
    const key = groupKey(result)
    const order = orderValue(result.group_order ?? result.plan_order)
    const previous = groups.get(key)
    if (!previous || order < previous.order) groups.set(key, { key, name: groupName(result), order })
  }
  return [...groups.values()].sort((a, b) => a.order - b.order || a.name.localeCompare(b.name))
})
watch(availableGroups, groups => {
  const requestedGroup = requestedGroupID.value > 0 ? `group:${requestedGroupID.value}` : ''
  if (requestedGroup && groups.some(group => group.key === requestedGroup)) activeGroup.value = requestedGroup
  else if (!groups.some(group => group.key === activeGroup.value)) activeGroup.value = groups[0]?.key || ''
}, { immediate: true })
const availableModels = computed(() => [...new Set(displayResults.value.map(result => result.model_id).filter((model): model is string => Boolean(model)))].sort())
const projectedVotes = computed(() => {
  const records = new Map(regularResults.value.map(result => [result.id, result]))
  return visibleVoteResults.value.map(item => {
    const current = records.get(item.result.id)
    return current ? { ...item, result: { ...item.result, group_id: current.group_id, group_name: current.group_name, group_order: current.group_order, plan_order: current.plan_order } } : item
  })
})
const filteredVotes = computed(() => projectedVotes.value.filter(item => groupKey(item.result) === activeGroup.value && (!modelFilter.value || item.result.model_id === modelFilter.value)))
const filteredResults = computed(() => sortResults(regularResults.value.filter(result => groupKey(result) === activeGroup.value && (!modelFilter.value || result.model_id === modelFilter.value))))
const schedulingState = (result: TestResult) => result.account_schedulable
  ?? (isHiddenPausedAccount(result) ? false : undefined)
const isSkippedResult = (result: TestResult) => result.status === 'failed' && (result.error_message || '').startsWith('检测跳过：')
const resultRank = (result: TestResult) => {
  if (isSkippedResult(result)) return 1
  if (['failed', 'error'].includes(result.status) || result.output_model_check?.verdict === 'fail') return 2
  if (!isSuccessful(result) || (result.output_kind === 'model_check' && result.output_model_check?.verdict !== 'pass')) return 1
  return 0
}
const accountResults = computed<TestAccountSeries[]>(() => {
  const accounts = new Map<string, { key: string; name: string; groupName: string; order: number; tests: Map<string, TestSeries> }>()
  for (const result of filteredResults.value) {
    const key = targetKey(result)
    let account = accounts.get(key)
    if (!account) {
      account = { key, name: t('tests.groupCheck'), groupName: groupName(result), order: exposesAccount(result) ? (result.result_set_id ?? result.account_id)! : -1, tests: new Map() }
      accounts.set(key, account)
    }
    const seriesKey = testKey(result)
    let series = account.tests.get(seriesKey)
    if (!series) {
      series = { key: seriesKey, name: testName(result), order: orderValue(result.test_order ?? result.test_definition?.sort_order), latest: result, results: [] }
      account.tests.set(seriesKey, series)
    }
    series.results.push(result)
  }
  return [...accounts.values()].map(account => {
    const tests = [...account.tests.values()].sort((a, b) => a.order - b.order || a.name.localeCompare(b.name) || a.key.localeCompare(b.key))
    const states = tests.map(test => schedulingState(test.latest))
    const schedulable = states.includes(false) ? false : states.includes(true) ? true : undefined
    const rank = schedulable === false ? 3 : Math.max(...tests.map(test => resultRank(test.latest)))
    return { ...account, tests, schedulable, rank }
  }).sort((a, b) => a.rank - b.rank || a.order - b.order)
})

const isSuccessful = (result: TestResult) => ['success', 'passed'].includes(result.status)
const historyAnchor = (series: TestSeries) => series.latest
const seriesIdentityFor = (account: Pick<TestAccountSeries, 'key'>, test: TestSeries) => `${account.key}|${groupKey(test.latest)}|${test.key}`
const selectedSeries = computed(() => {
  for (const account of accountResults.value) {
    for (const test of account.tests) if (seriesIdentityFor(account, test) === selectedSeriesId.value) return { account, test }
  }
  return null
})
const selectedResult = computed(() => {
  const selection = selectedSeries.value
  if (!selection) return undefined
  return selection.test.results.find(result => result.id === selectedResultId.value)
    ?? historyResults.value.find(result => result.id === selectedResultId.value)
    ?? (selection.test.latest.id === selectedResultId.value ? selection.test.latest : undefined)
})
watch(selectedResult, result => {
  if (previewTarget.value) previewTarget.value = result?.id === previewTarget.value.id ? result : null
})
const accountResultIds = computed(() => new Set(accountResults.value.flatMap(account => account.tests.map(test => test.latest.id))))
const standaloneVotes = computed(() => filteredVotes.value.filter(item => !accountResultIds.value.has(item.result.id)))
const workspaceHistoryItems = computed(() => {
  if (!selectedSeries.value) return []
  const records = new Map<number, TestResult>()
  for (const result of [...historyResults.value, ...selectedSeries.value.test.results]) records.set(result.id, result)
  return sortResults([...records.values()])
})
const selectedHistoryKey = computed(() => selectedSeries.value
  ? `${selectedSeriesId.value}|${selectedSeries.value.test.latest.plan_id ?? ''}|${historyAnchor(selectedSeries.value.test)?.id ?? ''}` : '')
let loadedSelection = ''
watch(selectedHistoryKey, key => {
  historyRequest++
  const selection = selectedSeries.value
  const identity = selection ? `${selectedSeriesId.value}|${selection.test.latest.plan_id ?? ''}` : ''
  if (identity !== loadedSelection) historyResults.value = []
  loadedSelection = identity
  historyBeforeId.value = undefined
  historyLoadedSeriesId.value = ''
  historyError.value = ''
  historyLoading.value = false
  adminHistoryLimit = 20
  if (key && selection && historyAnchor(selection.test)) void loadHistory(true)
}, { flush: 'post' })
watch(accountResults, accounts => {
  const current = accounts.flatMap(account => account.tests.map(test => ({ account, test })))
  if (!current.length) {
    if (selectedSeriesId.value) clearSelectedSeries()
    return
  }
  const matching = current.find(({ account, test }) => seriesIdentityFor(account, test) === selectedSeriesId.value)
  if (!matching) {
    const mostRecent = [...current].sort((a, b) => resultTimestamp(b.test.latest) - resultTimestamp(a.test.latest)
      || a.test.order - b.test.order || a.account.order - b.account.order)[0]
    selectSeries(mostRecent.account, mostRecent.test)
    return
  }
  // Keep following the newest result while the user is viewing the current
  // latest item. An explicitly selected historical record stays selected.
  if (followLatestSelection.value || (!matching.test.results.some(result => result.id === selectedResultId.value)
    && !historyResults.value.some(result => result.id === selectedResultId.value))) {
    selectedResultId.value = matching.test.latest.id
  }
}, { flush: 'post' })
const clearSelectedSeries = () => {
  historyRequest++
  selectedSeriesId.value = ''
  selectedResultId.value = undefined
  followLatestSelection.value = true
  historyLoadedSeriesId.value = ''
  historyResults.value = []
  historyBeforeId.value = undefined
  historyError.value = ''
  historyLoading.value = false
}
const loadHistory = async (refresh = false) => {
  const selection = selectedSeries.value
  const anchor = selection && historyAnchor(selection.test)
  if (!selection || !anchor || historyLoading.value || (!refresh && historyLoadedSeriesId.value === selectedSeriesId.value && !historyBeforeId.value)) return
  const request = ++historyRequest
  historyLoading.value = true
  historyError.value = ''
  try {
    // Administrator history can include results from before a group move.
    const planID = anchor.plan_id
    const page = requestedGroupID.value > 0
      ? await testResultsAPI.history(anchor.id, historyBeforeId.value, requestedGroupID.value)
      : auth.isAdmin && planID != null ? await (async () => {
      const rows = await listAdminResults(planID, adminHistoryLimit + 1)
      const items = sortResults(rows.filter(result => isSuccessful(result) && (exposesAccount(anchor) || groupKey(result) === groupKey(anchor))
        && targetKey(result) === targetKey(anchor) && testKey(result) === testKey(anchor)))
      return { items: items.slice(0, adminHistoryLimit), next_before_id: items.length > adminHistoryLimit ? items[adminHistoryLimit - 1].id : undefined }
    })() : await testResultsAPI.history(anchor.id, historyBeforeId.value)
    if (request !== historyRequest) return
    const records = new Map(historyResults.value.map(result => [result.id, result]))
    for (const result of page.items) records.set(result.id, result)
    historyResults.value = sortResults([...records.values()])
    historyBeforeId.value = page.next_before_id
    historyLoadedSeriesId.value = selectedSeriesId.value
    if (auth.isAdmin) adminHistoryLimit += 20
  } catch (error) {
    if (request !== historyRequest) return
    const failure = error as { status?: number; response?: { status?: number } }
    const status = failure?.status ?? failure?.response?.status
    if (status === 403 || status === 404) {
      historyResults.value = []
      historyBeforeId.value = undefined
      selectedResultId.value = selection.test.latest.id
      followLatestSelection.value = true
      historyError.value = extractApiErrorMessage(error, t('tests.loadFailed'))
      historyLoadedSeriesId.value = ''
      void load(true)
    } else historyError.value = extractApiErrorMessage(error, t('tests.loadFailed'))
  } finally {
    if (request === historyRequest) historyLoading.value = false
  }
}
const selectSeries = (account: TestAccountSeries, series: TestSeries) => {
  selectedSeriesId.value = seriesIdentityFor(account, series)
  selectedResultId.value = series.latest.id
  followLatestSelection.value = true
}
const selectResult = async (result: TestResult) => {
  selectedResultId.value = result.id
  followLatestSelection.value = result.id === selectedSeries.value?.test.latest.id
  await nextTick()
  resultGallery.value?.querySelector<HTMLElement>(`[data-result-card="${result.id}"]`)?.scrollIntoView?.({ block: 'nearest', behavior: 'smooth' })
}
const closeHistory = clearSelectedSeries
const voteFor = (result: TestResult) => filteredVotes.value.find(item => item.result.id === result.id)
const statusClass = (result: TestResult) => result.status === 'running' || result.status === 'pending' || isSkippedResult(result) ? 'badge badge-warning'
  : !isSuccessful(result) ? 'badge badge-danger'
  : result.output_kind === 'model_check' ? result.output_model_check?.verdict === 'pass' ? 'badge badge-success' : result.output_model_check?.verdict === 'fail' ? 'badge badge-danger' : 'badge badge-warning' : 'badge badge-success'
const statusLabel = (result: TestResult) => result.status === 'running' || result.status === 'pending' ? t('tests.running')
  : isSkippedResult(result) ? t('tests.skipped')
  : !isSuccessful(result) ? t('tests.failed')
  : result.output_kind === 'model_check' ? t(`tests.modelCheck.${result.output_model_check?.verdict || 'unknown'}`) : t('tests.completed')
const modelLabel = (result: TestResult) => `${result.model_id || '-'}${result.reasoning_effort ? ` · ${t('tests.reasoningEffort')}: ${result.reasoning_effort}` : ''}`
const formatDate = (value?: string) => value ? new Date(value).toLocaleString() : '-'
const onGroupKeydown = async (event: KeyboardEvent, key: string) => {
  const groups = availableGroups.value
  const index = groups.findIndex(group => group.key === key)
  let next = index
  if (event.key === 'ArrowRight') next = (index + 1) % groups.length
  else if (event.key === 'ArrowLeft') next = (index - 1 + groups.length) % groups.length
  else if (event.key === 'Home') next = 0
  else if (event.key === 'End') next = groups.length - 1
  else return
  event.preventDefault()
  activeGroup.value = groups[next].key
  await nextTick()
  document.getElementById(tabId(activeGroup.value))?.focus()
}
const refreshAfterDecision = () => {
  return load(true)
}
const reconcileOpenHistory = () => {
  if (!selectedSeriesId.value) return
  const selection = selectedSeries.value
  if (!selection) {
    closeHistory()
    return
  }
  const anchor = historyAnchor(selection.test)
  if (!anchor) return
  // A group move or revoked result access invalidates every cached page and
  // any in-flight history response, including when someone else moved it.
  const visible = displayResults.value.some(result => result.plan_id === anchor.plan_id && targetKey(result) === targetKey(anchor)
    && testKey(result) === testKey(anchor) && groupKey(result) === groupKey(anchor))
  if (!visible) closeHistory()
}
const visibleVote = (item: TestVoteResult) => item.voting.enabled && isSuccessful(item.result)
  && item.result.account_id != null && item.result.target_mode !== 'group' && item.result.output_kind !== 'statistics'
const castVote = async (id: number, vote: TestVote) => {
  if (votingResultIds.value.has(id)) return
  const current = voteResults.value.find(item => item.result.id === id)
  if (!current?.voting.open) return
  dataRevision++
  votingResultIds.value.add(id)
  delete voteErrors.value[id]
  try {
    await testResultsAPI.vote(id, vote)
  } catch (error) {
    if (!unmounted) {
      voteErrors.value[id] = extractApiErrorMessage(error, t('tests.voting.voteFailed'))
      // A closed or revoked result may disappear during the refresh below.
      app.showError(voteErrors.value[id])
    }
  } finally {
    dataRevision++
    if (!unmounted) {
      // Voting can move an account to a group the user cannot access. Never
      // retain an old result, an open history, or the POST response as proof
      // of continued access; reload both public projections after each vote.
      allResults.value = []
      voteResults.value = []
      closeHistory()
      previewTarget.value = null
      await load(true)
    }
    votingResultIds.value.delete(id)
  }
}
const load = (force = false): Promise<void> => {
  if (unmounted) return Promise.resolve()
  if (loadPromise && !force) return loadPromise
  loading.value = true
  const revision = ++dataRevision
  const administrator = auth.isAdmin
  const rawGroupID = Array.isArray(route.query.group_id) ? route.query.group_id[0] : route.query.group_id
  const groupID = rawGroupID && /^\d+$/.test(rawGroupID) ? Number(rawGroupID) : 0
  const request = Promise.allSettled([
    groupID > 0 ? testResultsAPI.listForGroup(groupID) : testResultsAPI.list(3),
    testResultsAPI.votes(),
    administrator ? listReviews() : Promise.resolve<TestAdminReview[]>([]),
  ]).then(([results, votes, reviews]) => {
    if (unmounted || revision !== dataRevision) return
    // Publish one complete refresh together. An old voting or review response
    // must never keep the previous group's tab alongside freshly moved results.
    allResults.value = results.status === 'fulfilled'
      ? sortResults(results.value.filter(result => ['success', 'passed', 'failed', 'error', 'pending', 'running'].includes(result.status))) : []
    voteResults.value = votes.status === 'fulfilled' ? votes.value.filter(visibleVote) : []
    adminReviews.value = administrator && auth.isAdmin && reviews.status === 'fulfilled' ? reviews.value : []
    votesLoadError.value = votes.status === 'rejected' ? extractApiErrorMessage(votes.reason, t('tests.voting.loadFailed')) : ''
    reviewError.value = administrator && reviews.status === 'rejected' ? extractApiErrorMessage(reviews.reason, t('tests.adminReview.loadFailed')) : ''
    if (results.status === 'rejected') app.showError(extractApiErrorMessage(results.reason, t('tests.loadFailed')))
    reconcileOpenHistory()
  }).finally(() => {
    if (loadPromise === request) {
      loading.value = false
      loadPromise = undefined
    }
  })
  loadPromise = request
  return request
}
watch(() => route.query.group_id, () => { void load(true) })
const refreshVisiblePage = () => {
  if (!document.hidden) void load(true)
}
let resultTimer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  void load()
  resultTimer = setInterval(() => { if (!document.hidden) void load() }, 5000)
  document.addEventListener('visibilitychange', refreshVisiblePage)
  window.addEventListener('focus', refreshVisiblePage)
})
onUnmounted(() => {
  unmounted = true
  dataRevision++
  clearInterval(resultTimer)
  document.removeEventListener('visibilitychange', refreshVisiblePage)
  window.removeEventListener('focus', refreshVisiblePage)
  historyRequest++
})
</script>
