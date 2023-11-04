import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
    thresholds: {
        http_req_duration: [{
            threshold: "p(99) < 150",
            abortOnFail: true
        }],
        http_req_failed: [
            {
                threshold: "rate<0.01",
                abortOnFail: true
            }
        ]

    },
    stages: [
        { duration: "10m", target: 100000 },
    ],
};

export default function () {

    let res = http.get("http://localhost:8080/");

    check(res, { "status was 200": (r) => r.status == 200 });
    sleep(1);
}