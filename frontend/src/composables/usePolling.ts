import { onMounted, onUnmounted } from 'vue'

export function usePolling(callback: () => void, interval = 30000) {
  let timer: number | null = null

  function start() {
    stop()
    timer = window.setInterval(callback, interval)
  }

  function stop() {
    if (timer !== null) {
      clearInterval(timer)
      timer = null
    }
  }

  onMounted(() => {
    callback()
    start()
  })

  onUnmounted(() => {
    stop()
  })

  return { start, stop }
}
