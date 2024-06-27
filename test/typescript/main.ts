import * as empty from "gen/empty";
import * as message from "gen/message";
import * as enumpb from "gen/enumpb";
import * as nested from "gen/nested";
import * as repeated from "gen/repeated";
import * as oneof from "gen/oneof";
import * as optional from "gen/optional";
import * as importing from "gen/importing";
import { Jsonif, getType, fromJson, toJson } from "gen/jsonif";

function assertEqual<T>(a: T, b: T) {
    if (a !== b) {
        throw new Error(`Expected ${a} to equal ${b}`);
    }
}

function identify<T extends number | string | boolean | Jsonif<T>>(v: T): T {
    const vs = toJson(v);
    const r = fromJson<T>(vs, getType(v));
    const rs = toJson(r);
    assertEqual(vs, rs);
    return r as T;
}

function testEmpty() {
  var a = new empty.Test();
  a = identify(a);
}

function testMessage() {
  var a = new message.Person();
  assertEqual(a.name, "");
  assertEqual(a.flag, false);
  a = identify(a);
  assertEqual(a.name, "");
  assertEqual(a.flag, false);

  a.name = "foo";
  a.flag = true;
  a = identify(a);
  assertEqual(a.name, "foo");
  assertEqual(a.flag, true);

  fromJson<message.Person>(toJson(a), message.Person);
}

function testEnumpb() {
  var a = enumpb.Data.FOO;
  a = identify(a);
  assertEqual(a, enumpb.Data.FOO);

  a = enumpb.Data.BAR;
  a = identify(a);
  assertEqual(a, enumpb.Data.BAR);
}

function testNested() {
  var a = new nested.Test2();
  assertEqual(a.nested_message.name, "");
  assertEqual(a.nested_enum, nested.Test_NestedEnum.FOO);
  assertEqual(a.test.nested_message.name, "");
  assertEqual(a.test.nested_enum, nested.Test_NestedEnum.FOO);
  a = identify(a);
  assertEqual(a.nested_message.name, "");
  assertEqual(a.nested_enum, nested.Test_NestedEnum.FOO);
  assertEqual(a.test.nested_message.name, "");
  assertEqual(a.test.nested_enum, nested.Test_NestedEnum.FOO);

  a.nested_message.name = "foo";
  a.nested_enum = nested.Test_NestedEnum.BAR;
  a.test.nested_message.name = "bar";
  a.test.nested_enum = nested.Test_NestedEnum.HOGE;
  a = identify(a);
  assertEqual(a.nested_message.name, "foo");
  assertEqual(a.nested_enum, nested.Test_NestedEnum.BAR);
  assertEqual(a.test.nested_message.name, "bar");
  assertEqual(a.test.nested_enum, nested.Test_NestedEnum.HOGE);
}

function testRepeated() {
  var a = new repeated.Test();
  assertEqual(a.a.length, 0);
  assertEqual(a.b.length, 0);
  assertEqual(a.c.length, 0);
  assertEqual(a.d.length, 0);
  a = identify(a);
  assertEqual(a.a.length, 0);
  assertEqual(a.b.length, 0);
  assertEqual(a.c.length, 0);
  assertEqual(a.d.length, 0);

  a.a.push(1);
  a.b.push("foo");
  a.c.push(repeated.Enum.BAR);
  a.d.push(new repeated.Message({name: "bar"}));
  a = identify(a);
  assertEqual(a.a.length, 1);
  assertEqual(a.b.length, 1);
  assertEqual(a.c.length, 1);
  assertEqual(a.d.length, 1);
  assertEqual(a.a[0], 1);
  assertEqual(a.b[0], "foo");
  assertEqual(a.c[0], repeated.Enum.BAR);
  assertEqual(a.d[0].name, "bar");
}

function testOneof() {
  var a = new oneof.Test();
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.NOT_SET);
  a = identify(a);
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.NOT_SET);

  a.setA(1);
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.kA);
  assertEqual(a.a, 1);
  a = identify(a);
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.kA);
  assertEqual(a.a, 1);

  a.setB("foo");
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.kB);
  assertEqual(a.b, "foo");
  a = identify(a);
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.kB);
  assertEqual(a.b, "foo");

  a.setC(oneof.Enum.BAR);
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.kC);
  assertEqual(a.c, oneof.Enum.BAR);
  a = identify(a);
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.kC);
  assertEqual(a.c, oneof.Enum.BAR);

  a.setD(new oneof.Message({name: "bar"}));
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.kD);
  assertEqual(a.d.name, "bar");
  a = identify(a);
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.kD);
  assertEqual(a.d.name, "bar");

  a.clearC();
  assertEqual(a.d.name, "bar");
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.kD);
  a.clearD();
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.NOT_SET);

  a.setA(10);
  a.clearTestOneof();
  assertEqual(a.test_oneof_case, oneof.Test_TestOneofCase.NOT_SET);
}

function testOptional() {
  var a = new optional.Test();
  assertEqual(a.a, null);
  assertEqual(a.b, null);
  assertEqual(a.c, null);
  assertEqual(a.d, null);
  a = identify(a);
  assertEqual(a.a, null);
  assertEqual(a.b, null);
  assertEqual(a.c, null);
  assertEqual(a.d, null);

  a.a = 1;
  assertEqual(a.a, 1);
  a = identify(a);
  assertEqual(a.a, 1);

  a.b = "foo";
  assertEqual(a.b, "foo");
  a = identify(a);
  assertEqual(a.b, "foo");

  a.c = optional.Enum.BAR;
  assertEqual(a.c, optional.Enum.BAR);
  a = identify(a);
  assertEqual(a.c, optional.Enum.BAR);

  a.d = new optional.Message({name: "bar"});
  assertEqual(a.d.name, "bar");
  a = identify(a);
  assertEqual(a.d!.name, "bar");

  a.a = null;
  assertEqual(a.a, null);
  a.b = null;
  assertEqual(a.b, null);
  a.c = null;
  assertEqual(a.c, null);
  a.d = null;
  assertEqual(a.d, null);
}

function testImporting() {
  var a = new importing.Test();
  assertEqual(a.t.nanos, 0);
  a = identify(a);
  assertEqual(a.t.nanos, 0);
}

testEmpty();
testMessage();
testEnumpb();
testNested();
testRepeated();
testOneof();
testOptional();
testImporting();
