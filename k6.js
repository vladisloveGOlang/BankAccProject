import http from 'k6/http';
import { sleep } from 'k6';
import { check } from 'k6';
import { randomIntBetween, randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// @todo: mv to testify

export const options = {
    vus: 20,
    duration: '20s',
    ext: {
        loadimpact: {
            distribution: {
                distributionLabel2: { loadZone: 'amazon:de:frankfurt', percent: 100 },
            },
            // Project: crm
            projectID: 3675081,
            // Test runs with the same name groups test runs together.
            name: 'Test Tasks'
        }
    }
};

const binFile = open('./testify/assets/photo-2.jpg', 'b');

export default function () {

    const cookies = {
        TOKEN: {
            value: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiMTIzZjAwYjAtMzc2OS00M2RlLTgzMjUtNTQzNTRiYTdlYTg2IiwiZW1haWwiOiJuaWdodHNvbmdAb3Zpb3ZpLnNpdGUiLCJuYW1lIjoic29uZyIsImlzX3ZhbGlkIjpmYWxzZSwiaXNzIjoidG9kbyIsInN1YiI6InVzZXIiLCJhdWQiOlsidG9kbyJdLCJleHAiOjE3MDQ3NjM1MDQsImlhdCI6MTcwNDc2MzQ5OX0.T5mt-NO5dn8fHoZy3B0e5e1KU_oX_-2DXrYXohFxgpw&eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiMTIzZjAwYjAtMzc2OS00M2RlLTgzMjUtNTQzNTRiYTdlYTg2IiwiZW1haWwiOiJuaWdodHNvbmdAb3Zpb3ZpLnNpdGUiLCJuYW1lIjoic29uZyIsImlzX3ZhbGlkIjpmYWxzZSwiaXNzIjoidG9kbyIsInN1YiI6InJlZnJlc2giLCJhdWQiOlsidG9kbyJdLCJleHAiOjE3MDQ4MDY2OTksImlhdCI6MTcwNDc2MzQ5OX0.EBqfZrJvgY5_4HlTvkgpIimu0-131t2RZjbTjimLz9M',
            replace: true,
        },
    };

    const params = {
        headers: {

        },
        cookies,
    };

    // const data = {
    //     file: http.file(binFile, 'photo-5.jpg'),
    // };

    // const res = http.patch('http://localhost:8080/profile/photo', data, params);


    const status = randomIntBetween(0, 1)
    const federationUUID = 'cb06b506-f46f-4bf4-9edb-2b12b1367681'
    const projectUUID = '8784f657-0f21-459e-b305-29a0cbda796e'
    const limit = 25
    const name = randomString(1)
    const isEpic = status === 1 ? true : false

    const url = `https://oviovi.site/api/task?federation_uuid=${federationUUID}&is_epic=${isEpic}&project_uuid=${projectUUID}&status=${status}&limit=${limit}&name=${name}`

    const res = http.get(url, params);

    check(res, {
        'is status 200': (r) => r.status === 200,
    });

    check(res, {
        'is found': (r) => JSON.parse(res.body).count > 0,
    });
}
