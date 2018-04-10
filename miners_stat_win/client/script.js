

var app = new Vue({
    el: '#app',
    data: () => {
        return {
            showIp: false,
            table: null,
            error: false
        }
    },
    mounted(a) {
        loadapi()
        setInterval(loadapi, 10000)

    },
    methods: {
        reboot(ip) {
            var psswd = prompt('Enter PSSWD')
            var url = 'http://localhost:8089/miners/reboot/' + ip
            axios.get(url + '/' + psswd.trim())
                .then((res) => {
                    alert('OK')
                })
                .catch((err) => {
                    alert(err)
                })
        },

    }
})

Vue.filter('log', (d) => {
    console.log(d)
})

function loadapi() {
    var miner = window.location.search.replace('?', '')
    var url = 'http://localhost:8089/miners'
    if (miner.length) {
        url += '/' + miner
    }
    axios.get(url)
        .then((res) => {
            app.$data.error = false
            app.$data.table = res.data
            var play = false
            Object.keys(res.data).forEach((key) => {
                Object.keys(res.data[key]).forEach((worker) => {
                    var v = res.data[key][worker]
                    if (!v.online) {
                        play = true
                    }
                    if (v.status) {
                        v.status.e_gpu_hashrate.forEach((hr) => {
                            if (hr === 0) {
                                play = true
                            }

                        })
                    }

                })
            })
            if (play) {
                audio.currentTime = 0
                audio.play()
            } else {
                audio.pause()
            }
        })
        .catch((err) => {
            app.$data.error = true
        })

}

Vue.filter('formatHashrate', formatHashrate)
function formatHashrate(
    hashrate,
    fixed = 2,
    del = ' ',
    prefix = 'h'
) {
    let i = 0
    const units = [
        prefix,
        'K' + prefix,
        'M' + prefix,
        'G' + prefix,
        'T' + prefix,
        'P' + prefix
    ]
    while (hashrate >= 1000) {
        hashrate = hashrate / 1000
        i++
    }

    if (!hashrate) {
        hashrate = 0
    }

    return hashrate.toFixed(fixed) + del + units[i]
}


Vue.filter('formattime', formattime)

function formattime(
    time
) {
    var mins = time % 60;
    var hours = (time - mins) / 60;
    if (mins < 10) mins = '0' + mins;
    if (hours < 10) hours = '0' + hours;
    var rezult = hours + ':' + mins; // нужный вам формат
    return rezult
}