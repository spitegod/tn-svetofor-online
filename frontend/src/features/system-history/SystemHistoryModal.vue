<script setup lang="ts">
import { ChevronUp, Clock3, ExternalLink, Globe2, House, Info, RefreshCw, X } from '@lucide/vue'
import genericFileIcon from 'bootstrap-icons/icons/file-earmark.svg'
import type { SystemDocumentRow } from '@/shared/api/types'

defineProps<{
  system: SystemDocumentRow | null
  rows: SystemDocumentRow[]
  isLoading: boolean
  error: string
  isHistoryOpen: boolean
  attachmentFileKind: (name: string) => string
  attachmentFileIcon: (name: string) => string
  attachmentURL: (row: SystemDocumentRow) => string
}>()

const emit = defineEmits<{
  close: []
  'update:isHistoryOpen': [value: boolean]
}>()
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="system" class="modal-overlay" @click="emit('close')">
        <section class="system-history-card" aria-label="История изменений системы" @click.stop>
          <button class="modal-close-button" type="button" aria-label="Закрыть" @click="emit('close')">
            <X :size="27" :stroke-width="2" aria-hidden="true" />
          </button>

          <header class="system-history-header">
            <span class="system-history-header__icon" aria-hidden="true">
              <House :size="31" :stroke-width="1.8" />
            </span>
            <div class="system-history-header__content">
              <div class="system-history-header__title">
                <h2>{{ system.systemName }}</h2>
                <a class="system-history-source" :href="system.systemUrl || 'https://nav.tn.ru/systems/'" target="_blank" rel="noreferrer" aria-label="Открыть на nav.tn.ru">
                  <Globe2 :size="19" :stroke-width="2" aria-hidden="true" />
                </a>
              </div>
              <p>История изменений распоряжений и связанных документов</p>
            </div>
          </header>

          <div v-if="isLoading" class="system-history-loading">
            <RefreshCw :size="20" :stroke-width="1.8" aria-hidden="true" />
            Загрузка истории…
          </div>
          <p v-else-if="error" class="table-message table-message--error">{{ error }}</p>

          <template v-else-if="rows.length">
            <section class="history-current-section">
              <div class="history-columns" aria-hidden="true">
                <span>Распоряжение</span>
                <span>Комментарий</span>
                <span>Документ</span>
              </div>

              <article v-for="row in rows.slice(0, 1)" :key="`history-current-${row.id}`" class="history-entry history-entry--current">
                <div class="history-entry__order"><strong>{{ row.orderName }}</strong></div>
                <p :class="{ 'history-comment--empty': !row.comment }">{{ row.comment || 'Комментарий не добавлен' }}</p>
                <a v-if="row.attachmentName" class="history-document" :href="attachmentURL(row)" target="_blank" rel="noreferrer">
                  <span class="history-document__type-icon" :class="`history-document__type-icon--${attachmentFileKind(row.attachmentName)}`" aria-hidden="true">
                    <img :src="attachmentFileIcon(row.attachmentName)" alt="" />
                  </span>
                  <span>{{ row.attachmentName }}</span>
                  <ExternalLink :size="17" :stroke-width="1.8" aria-hidden="true" />
                </a>
                <span v-else class="history-document history-document--empty">
                  <span class="history-document__empty-icon" aria-hidden="true"><img :src="genericFileIcon" alt="" /></span>
                  <span>Документ не прикреплён</span>
                </span>
              </article>
            </section>

            <button v-if="rows.length > 1" class="history-toggle" type="button" @click="emit('update:isHistoryOpen', !isHistoryOpen)">
              <span class="history-toggle__line" aria-hidden="true" />
              <ChevronUp :class="{ 'is-collapsed': !isHistoryOpen }" :size="20" :stroke-width="2" aria-hidden="true" />
              {{ isHistoryOpen ? 'Скрыть историю изменений' : 'Показать историю изменений' }}
              <span class="history-toggle__line" aria-hidden="true" />
            </button>

            <Transition name="history-more">
              <section v-if="isHistoryOpen && rows.length > 1" class="history-archive">
                <header>
                  <span aria-hidden="true"><Clock3 :size="18" :stroke-width="1.8" /></span>
                  <strong>История изменений</strong>
                </header>
                <div class="history-timeline">
                  <article v-for="row in rows.slice(1)" :key="`history-${row.id}`" class="history-entry history-entry--past">
                    <div class="history-entry__order"><strong>{{ row.orderName }}</strong></div>
                    <p :class="{ 'history-comment--empty': !row.comment }">{{ row.comment || 'Комментарий не добавлен' }}</p>
                    <a v-if="row.attachmentName" class="history-document" :href="attachmentURL(row)" target="_blank" rel="noreferrer">
                      <span class="history-document__type-icon" :class="`history-document__type-icon--${attachmentFileKind(row.attachmentName)}`" aria-hidden="true">
                        <img :src="attachmentFileIcon(row.attachmentName)" alt="" />
                      </span>
                      <span>{{ row.attachmentName }}</span>
                      <ExternalLink :size="17" :stroke-width="1.8" aria-hidden="true" />
                    </a>
                    <span v-else class="history-document history-document--empty">
                      <span class="history-document__empty-icon" aria-hidden="true"><img :src="genericFileIcon" alt="" /></span>
                      <span>Документ не прикреплён</span>
                    </span>
                  </article>
                </div>
              </section>
            </Transition>

            <footer class="system-history-footer">
              <Info :size="17" :stroke-width="1.8" aria-hidden="true" />
              Самые новые изменения отображаются первыми.
            </footer>
          </template>

          <p v-else class="system-history-empty">История изменений пока отсутствует</p>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>
