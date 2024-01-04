import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
    vus: 10, // number of virtual users
    duration: '30s', // test duration
};

export default function () {
    http.get('http://localhost:8080/test');
    sleep(1);
};