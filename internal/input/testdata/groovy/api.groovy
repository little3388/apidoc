// SPDX-License-Identifier: MIT

package testdata;

public class test1{

    // @api POST /users/login 登录
    // group users
    // tags: [t1,t2]
    //
    // request:
    //   description: request body
    //   content:
    //     application/json:
    //       schema:
    //         type: object
    //         properties:
    //           username:
    //             type: string
    //             description: 登录账号
    //           password:
    //             type: string
    //             description: 密码
    public void login() {
        System.out.println("/********** login");
        System.out.println("1123\\");
        System.out.println('''1123\\''');
        System.out.println('1123\\');
        // TODO
    }

    // 123
    // 123
    /* @api DELETE /users/login 注销登录
    group users
    tags: [t1,t2]

    request:
      description: request body
      content:
        application/json:
          schema:
            type: object
            properties:
              username:
                type: string
                description: 登录账号
              password:
                type: string
                description: 密码
*/
    public void logout() {
        System.out.println("logout **********/");
        // TODO
    }
}
