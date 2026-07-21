<script setup lang="ts">
import { ChevronDown, CircleCheck, Info, RefreshCw, TriangleAlert, X } from '@lucide/vue'

export type NavParserPanelViewModel = {
  [key: string]: any
}

const props = defineProps<{ model: NavParserPanelViewModel }>()

const {
  cancelNavParser,
  formatNavParserLogTime,
  formatNavParserRunDate,
  formatNavParserRunDuration,
  isNavParserCancelling,
  isNavParserHistoryOpen,
  isNavParserLogOpen,
  isNavParserSettingsOpen,
  isNavParsing,
  isNavSettingsSaving,
  navParseError,
  navParseFailedSystems,
  navParseMessage,
  navParseNotFound,
  navParserFallbackSearch,
  navParserIntervalDays,
  navParserLogsNewestFirst,
  navParserNextRunLabel,
  navParserProgress,
  navParserRequestTimeout,
  navParserRetryAttempts,
  navParserRetryDelay,
  navParserRunLogsNewestFirst,
  navParserRuns,
  navParserSourceLabel,
  navParserWorkerCount,
  navSettingsError,
  navSettingsMessage,
  openedNavParserRunId,
  runNavParser,
  saveNavParserSettings,
} = props.model
</script>

<template>
        <section class="settings-section parser-settings" aria-labelledby="parser-settings-title">
          <div class="parser-settings__main">
            <div class="parser-settings__identity">
              <span class="parser-settings__icon" aria-hidden="true">
                <RefreshCw :size="42" :stroke-width="1.8" />
              </span>
              <div class="parser-settings__content">
                <h1 id="parser-settings-title">Парсинг навигатора</h1>
                <div class="parser-settings__schedule">
                  <span>Автозапуск каждые <strong>{{ navParserIntervalDays }}</strong> дн.</span>
                  <small>{{ navParserNextRunLabel() }}</small>
                </div>
              </div>
            </div>
            <div class="parser-settings__controls">
              <button
                class="import-button parser-settings__button"
                :class="{ 'parser-settings__button--stop': isNavParsing }"
                type="button"
                :disabled="isNavParserCancelling"
                @click="isNavParsing ? cancelNavParser() : runNavParser()"
              >
                {{ isNavParserCancelling ? 'Останавливаем…' : isNavParsing ? 'Остановить парсинг' : 'Запустить парсер' }}
              </button>
            </div>
          </div>
          <section v-if="navParserProgress.running || navParserProgress.logs.length" class="parser-progress" aria-live="polite">
            <header class="parser-progress__header">
              <div>
                <span class="parser-progress__state" :class="{ 'is-running': navParserProgress.running }">
                  <RefreshCw v-if="navParserProgress.running" :size="15" :stroke-width="2" aria-hidden="true" />
                  <CircleCheck v-else-if="navParserProgress.percent === 100" :size="15" :stroke-width="2" aria-hidden="true" />
                  <Info v-else :size="15" :stroke-width="2" aria-hidden="true" />
                  {{ navParserProgress.stage }}
                </span>
                <strong>{{ navParserProgress.message }}</strong>
              </div>
              <b>{{ navParserProgress.percent }}%</b>
            </header>
            <div
              class="parser-progress__track"
              role="progressbar"
              :aria-valuenow="navParserProgress.percent"
              aria-valuemin="0"
              aria-valuemax="100"
              :aria-label="`Прогресс парсера: ${navParserProgress.percent}%`"
            >
              <span :style="{ width: `${navParserProgress.percent}%` }" />
            </div>
            <div class="parser-progress__stats">
              <span><small>Обработано</small><strong>{{ navParserProgress.processed }} / {{ navParserProgress.total }}</strong></span>
              <span><small>Найдено</small><strong>{{ navParserProgress.found }}</strong></span>
              <span><small>Обновлено</small><strong>{{ navParserProgress.updated }}</strong></span>
              <span><small>Не найдено</small><strong>{{ navParserProgress.notFound }}</strong></span>
              <span><small>Ошибки</small><strong>{{ navParserProgress.failed }}</strong></span>
            </div>
            <section class="parser-log">
              <button
                class="parser-log__toggle"
                type="button"
                :aria-expanded="isNavParserLogOpen"
                @click="isNavParserLogOpen = !isNavParserLogOpen"
              >
                <span>Журнал выполнения</span>
                <span class="parser-log__toggle-meta">
                  <small>{{ navParserProgress.logs.length }} записей</small>
                  <ChevronDown :size="16" :stroke-width="2" :class="{ 'is-open': isNavParserLogOpen }" aria-hidden="true" />
                </span>
              </button>
              <Transition name="parser-log-body">
                <div v-if="isNavParserLogOpen" class="parser-log__body">
                  <div class="parser-log__entries">
                    <p v-for="(entry, index) in navParserLogsNewestFirst" :key="`${entry.time}-${index}`" :class="`parser-log__entry parser-log__entry--${entry.level}`">
                      <time>{{ formatNavParserLogTime(entry.time) }}</time>
                      <i aria-hidden="true" />
                      <span>{{ entry.message }}</span>
                    </p>
                  </div>
                </div>
              </Transition>
            </section>
          </section>
          <section class="parser-options-section">
            <button
              class="parser-options__toggle"
              type="button"
              :aria-expanded="isNavParserHistoryOpen"
              @click="isNavParserHistoryOpen = !isNavParserHistoryOpen"
            >
              <span>Журнал запусков</span>
              <ChevronDown :class="{ 'is-open': isNavParserHistoryOpen }" :size="17" :stroke-width="1.9" aria-hidden="true" />
            </button>
            <Transition name="parser-options">
              <div v-if="isNavParserHistoryOpen" class="parser-options__shell">
                <section class="parser-history" aria-label="Журнал запусков">
                  <p v-if="!navParserRuns.length" class="parser-history__empty">Завершённых запусков пока нет.</p>
                  <div v-else class="parser-history__list">
              <article v-for="run in navParserRuns" :key="run.id" class="parser-history__run" :class="`is-${run.status}`">
                <button
                  class="parser-history__run-toggle"
                  type="button"
                  :aria-expanded="openedNavParserRunId === run.id"
                  @click="openedNavParserRunId = openedNavParserRunId === run.id ? null : run.id"
                >
                  <span class="parser-history__status" aria-hidden="true">
                    <CircleCheck v-if="run.status === 'completed'" :size="17" :stroke-width="2" />
                    <X v-else-if="run.status === 'canceled'" :size="17" :stroke-width="2" />
                    <TriangleAlert v-else :size="17" :stroke-width="2" />
                  </span>
                  <span class="parser-history__identity">
                    <strong>{{ formatNavParserRunDate(run.startedAt) }}</strong>
                    <small>{{ navParserSourceLabel(run.source) }} · {{ formatNavParserRunDuration(run) }}</small>
                  </span>
                  <span class="parser-history__summary">
                    <span>Обновлено <b>{{ run.updated }}</b></span>
                    <span>Не найдено <b>{{ run.notFound }}</b></span>
                    <span>Ошибки <b>{{ run.failed }}</b></span>
                  </span>
                  <ChevronDown :size="18" :stroke-width="2" :class="{ 'is-open': openedNavParserRunId === run.id }" aria-hidden="true" />
                </button>
                <Transition name="parser-log-body">
                  <div v-if="openedNavParserRunId === run.id" class="parser-log__body parser-history__log-body">
                    <div class="parser-log__entries">
                      <p v-for="(entry, index) in navParserRunLogsNewestFirst(run)" :key="`${run.id}-${entry.time}-${index}`" :class="`parser-log__entry parser-log__entry--${entry.level}`">
                        <time>{{ formatNavParserLogTime(entry.time) }}</time>
                        <i aria-hidden="true" />
                        <span>{{ entry.message }}</span>
                      </p>
                    </div>
                  </div>
                </Transition>
              </article>
                  </div>
                </section>
              </div>
            </Transition>
          </section>
          <section class="parser-options-section">
            <button
              class="parser-options__toggle"
              type="button"
              :aria-expanded="isNavParserSettingsOpen"
              @click="isNavParserSettingsOpen = !isNavParserSettingsOpen"
            >
              <span>Настройки</span>
              <ChevronDown :class="{ 'is-open': isNavParserSettingsOpen }" :size="17" :stroke-width="1.9" aria-hidden="true" />
            </button>
            <Transition name="parser-options">
              <div v-if="isNavParserSettingsOpen" class="parser-options__shell">
                <div class="parser-options">
                  <div class="parser-options__heading">
                    <div>
                      <strong>Расширенные настройки</strong>
                    </div>
                    <button class="parser-options__save" type="button" :disabled="isNavSettingsSaving" @click="saveNavParserSettings">
                      {{ isNavSettingsSaving ? 'Сохранение…' : 'Сохранить настройки' }}
                    </button>
                  </div>
                  <div class="parser-options__grid">
                    <label>
                      <span>Период обновления</span>
                      <small>Через сколько дней парсер сам запустится снова.</small>
                      <span class="parser-options__input"><input v-model.number="navParserIntervalDays" type="number" min="1" max="365" /><em>дней</em></span>
                    </label>
                    <label>
                      <span>Параллельные запросы</span>
                      <small>Сколько страниц nav.tn.ru загружать одновременно. Рекомендуемое значение — 4.</small>
                      <span class="parser-options__input"><input v-model.number="navParserWorkerCount" type="number" min="1" max="10" /><em>шт.</em></span>
                    </label>
                    <label>
                      <span>Тайм-аут запроса</span>
                      <small>Сколько ждать ответа сайта, прежде чем считать запрос неудачным. Рекомендуется 35 секунд.</small>
                      <span class="parser-options__input"><input v-model.number="navParserRequestTimeout" type="number" min="5" max="120" /><em>сек.</em></span>
                    </label>
                    <label>
                      <span>Количество попыток</span>
                      <small>Сколько раз повторить запрос, если сайт временно не ответил. Безопасное значение — 3.</small>
                      <span class="parser-options__input"><input v-model.number="navParserRetryAttempts" type="number" min="1" max="5" /><em>раз</em></span>
                    </label>
                    <label>
                      <span>Задержка между попытками</span>
                      <small>Пауза перед повторным запросом.</small>
                      <span class="parser-options__input"><input v-model.number="navParserRetryDelay" type="number" min="1" max="30" /><em>сек.</em></span>
                    </label>
                    <label class="parser-options__switch">
                      <span>
                        <strong>Резервный поиск</strong>
                        <small>Если система не найдена в общем каталоге, искать её отдельно через поиск nav.tn.ru.</small>
                      </span>
                      <input v-model="navParserFallbackSearch" type="checkbox" />
                    </label>
                  </div>
                </div>
              </div>
            </Transition>
          </section>
          <p v-if="navSettingsError" class="table-message table-message--error">{{ navSettingsError }}</p>
          <p v-else-if="navSettingsMessage" class="table-message table-message--success">{{ navSettingsMessage }}</p>
          <p v-if="navParseError" class="table-message table-message--error">{{ navParseError }}</p>
          <p v-else-if="navParseMessage" class="table-message table-message--success">{{ navParseMessage }}</p>
          <details v-if="navParseNotFound.length" class="parser-settings__not-found">
            <summary>Не найденные на nav.tn.ru системы ({{ navParseNotFound.length }})</summary>
            <ul>
              <li v-for="systemName in navParseNotFound" :key="systemName">{{ systemName }}</li>
            </ul>
          </details>
          <details v-if="navParseFailedSystems.length" class="parser-settings__not-found parser-settings__failed">
            <summary>Ошибки загрузки систем ({{ navParseFailedSystems.length }})</summary>
            <ul>
              <li v-for="systemName in navParseFailedSystems" :key="systemName">{{ systemName }}</li>
            </ul>
          </details>
        </section>
</template>
