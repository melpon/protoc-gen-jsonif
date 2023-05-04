#include <iostream>
#include <cassert>
#include <boost/json/src.hpp>

#include "empty.json.c.h"
#include "message.json.c.h"
#include "enumpb.json.c.h"
#include "nested.json.c.h"
#include "repeated.json.c.h"
#include "oneof.json.c.h"
#include "importing.json.c.h"
#include "bytes.json.c.h"
// #include "jsonfield.json.h"
// #include "optimistic.json.h"
// #include "discard_if_default.json.h"
// #include "no_serializer.json.h"

// template<class T>
// T identify(T v) {
//   auto vs = jsonif::to_json(v);
//   auto r = jsonif::from_json<T>(vs);
//   auto rs = jsonif::to_json(v);
//   assert(r == v);
//   assert(rs == vs);
//   return r;
// }

#define CONCAT(a, b) CONCAT_(a, b)
#define CONCAT_(a, b) a##b

#define TEST_IDENTIFY(t, v, u) \
  do { \
    std::string vs(CONCAT(t, _to_json_size)(v) - 1, '\0'); \
    CONCAT(t, _to_json)(v, &vs[0]); \
    t r; \
    CONCAT(t, _from_json)(vs.c_str(), &r); \
    std::string rs(CONCAT(t, _to_json_size)(v) - 1, '\0'); \
    CONCAT(t, _to_json)(v, &rs[0]); \
    assert(CONCAT(t, _is_equal)(&r, v)); \
    assert(rs == vs); \
    CONCAT(t, _copy)(&r, u); \
    CONCAT(t, _destroy)(&r); \
  } while (0)

void test_empty() {
  empty_Test a;
  empty_Test_init(&a);
  empty_Test b;
  empty_Test_init(&b);
  TEST_IDENTIFY(empty_Test, &a, &b);
}

void test_message() {
  message_Person a;
  message_Person_init(&a);
  assert(a.name == NULL);
  assert(a.name_len == 0);
  message_Person b;
  message_Person_init(&b);
  TEST_IDENTIFY(message_Person, &a, &b);
  assert(b.name == NULL);
  assert(b.name_len == 0);

  message_Person_set_name(&b, "foo");
  message_Person c;
  message_Person_init(&c);
  TEST_IDENTIFY(message_Person, &b, &c);
  assert(strcmp(c.name, "foo") == 0);
  assert(c.name_len == 3);

  message_Person_destroy(&a);
  message_Person_destroy(&b);
  message_Person_destroy(&c);
}

void test_enumpb() {
  enumpb_Data a = enumpb_Data_FOO;
  assert(a == enumpb_Data_FOO);
  a = enumpb_Data_BAR;
  assert(a == enumpb_Data_BAR);
}

void test_nested() {
  nested_nested_Test2 a;
  nested_nested_Test2_init(&a);
  assert(a.nested_message.name == NULL);
  assert(a.nested_message.name_len == 0);
  assert(a.nested_enum == nested_nested_Test_NestedEnum_FOO);
  assert(a.test.nested_message.name == NULL);
  assert(a.test.nested_enum == nested_nested_Test_NestedEnum_FOO);
  nested_nested_Test2 b;
  nested_nested_Test2_init(&b);
  TEST_IDENTIFY(nested_nested_Test2, &a, &b);
  assert(b.nested_message.name == NULL);
  assert(b.nested_message.name_len == 0);
  assert(b.nested_enum == nested_nested_Test_NestedEnum_FOO);
  assert(b.test.nested_message.name == NULL);
  assert(b.test.nested_message.name_len == 0);
  assert(b.test.nested_enum == nested_nested_Test_NestedEnum_FOO);

  nested_nested_Test_NestedMessage_set_name(&b.nested_message, "foo");
  nested_nested_Test2_set_nested_enum(&b, nested_nested_Test_NestedEnum_BAR);
  nested_nested_Test_NestedMessage_set_name(&b.test.nested_message, "bar");
  nested_nested_Test_set_nested_enum(&b.test, nested_nested_Test_NestedEnum_HOGE);
  nested_nested_Test2 c;
  nested_nested_Test2_init(&c);
  TEST_IDENTIFY(nested_nested_Test2, &b, &c);
  assert(strcmp(c.nested_message.name, "foo") == 0);
  assert(c.nested_message.name_len == 3);
  assert(c.nested_enum == nested_nested_Test_NestedEnum_BAR);
  assert(strcmp(c.test.nested_message.name, "bar") == 0);
  assert(c.test.nested_message.name_len == 3);
  assert(c.test.nested_enum == nested_nested_Test_NestedEnum_HOGE);
}

void test_repeated() {
  repeated_Test a;
  repeated_Test_init(&a);
  assert(a.a == nullptr);
  assert(a.a_len == 0);
  assert(a.b == nullptr);
  assert(a.b_len == 0);
  assert(a.c == nullptr);
  assert(a.c_len == 0);
  assert(a.d == nullptr);
  assert(a.d_len == 0);
  repeated_Test b;
  repeated_Test_init(&b);
  TEST_IDENTIFY(repeated_Test, &a, &b);
  assert(b.a == nullptr);
  assert(b.a_len == 0);
  assert(b.b == nullptr);
  assert(b.b_len == 0);
  assert(b.c == nullptr);
  assert(b.c_len == 0);
  assert(b.d == nullptr);
  assert(b.d_len == 0);

  repeated_Test_alloc_a(&b, 1);
  repeated_Test_set_a(&b, 0, 1);
  repeated_Test_set_b(&b, 0, "foo");
  repeated_Test_set_c(&b, 0, repeated_Enum_BAR);
  repeated_Message_set_name(&b.d[0], "bar");
  repeated_Test c;
  repeated_Test_init(&c);
  TEST_IDENTIFY(repeated_Test, &b, &c);
  assert(c.a_len == 1 && c.a[0] == 1);
  assert(c.b_len == 1 && c.b[0] == "foo" && c.b_lens[0] == 3);
  assert(c.c_len == 1 && c.c[0] == repeated_Enum_BAR);
  assert(c.d_len == 1 && strcmp(c.d[0].name, "bar") == 0);
}

void test_oneof() {
  oneof_Test a;
  oneof_Test_init(&a);
  oneof_Test b;
  oneof_Test_init(&b);
  assert(a.test_oneof_case == oneof_Test_TestOneofCase_NOT_SET);
  TEST_IDENTIFY(oneof_Test, &a, &b);
  assert(b.test_oneof_case == oneof_Test_TestOneofCase_NOT_SET);

  oneof_Test_set_a(&a, 1);
  assert(a.test_oneof_case == oneof_Test_TestOneofCase_kA);
  assert(a.a == 1);
  TEST_IDENTIFY(oneof_Test, &a, &b);
  assert(b.test_oneof_case == oneof_Test_TestOneofCase_kA);
  assert(b.a == 1);

  oneof_Test_set_b(&a, "foo");
  assert(a.test_oneof_case == oneof_Test_TestOneofCase_kB);
  assert(a.b_len == 3 && strcmp(a.b, "foo") == 0);
  TEST_IDENTIFY(oneof_Test, &a, &b);
  assert(b.test_oneof_case == oneof_Test_TestOneofCase_kB);
  assert(b.b_len == 3 && strcmp(b.b, "foo") == 0);

  oneof_Test_set_c(&a, oneof_Enum_BAR);
  assert(a.test_oneof_case == oneof_Test_TestOneofCase_kC);
  assert(a.c == oneof_Enum_BAR);
  TEST_IDENTIFY(oneof_Test, &a, &b);
  assert(b.test_oneof_case == oneof_Test_TestOneofCase_kC);
  assert(b.c == oneof_Enum_BAR);

  oneof_Message m;
  oneof_Message_init(&m);
  oneof_Message_set_name(&m, "bar");
  oneof_Test_set_d(&a, &m);
  oneof_Message_destroy(&m);
  assert(a.test_oneof_case == oneof_Test_TestOneofCase_kD);
  assert(a.d.name_len == 3 && strcmp(a.d.name, "bar") == 0);
  TEST_IDENTIFY(oneof_Test, &a, &b);
  assert(b.test_oneof_case == oneof_Test_TestOneofCase_kD);
  assert(b.d.name_len == 3 && strcmp(b.d.name, "bar") == 0);

  oneof_Test_clear_test_oneof_case(&a);
  oneof_Test_clear_test_oneof_case(&b);
  assert(a.test_oneof_case == oneof_Test_TestOneofCase_NOT_SET);
  assert(b.test_oneof_case == oneof_Test_TestOneofCase_NOT_SET);

  oneof_Test_destroy(&a);
  oneof_Test_destroy(&b);
}

void test_importing() {
  importing_Test a;
  importing_Test_init(&a);
  assert(a.t.nanos == 0);
  importing_Test b;
  importing_Test_init(&b);
  TEST_IDENTIFY(importing_Test, &a, &b);
  assert(b.t.nanos == 0);
}

void test_bytes() {
  std::string v("\x00\x01\x02\x03", 4);
  std::string v2 = u8"あいうえお";
  bytes_Test a;
  bytes_Test_init(&a);
  bytes_Test_set_data(&a, (const uint8_t*)v.data(), v.size());
  bytes_Test_alloc_rp_data(&a, 2);
  bytes_Test_set_rp_data(&a, 0, (const uint8_t*)v.data(), v.size());
  bytes_Test_set_rp_data(&a, 0, (const uint8_t*)v2.data(), v2.size());
  bytes_Test b;
  bytes_Test_init(&b);
  TEST_IDENTIFY(bytes_Test, &a, &b);
  assert(b.data_len == v.size() && memcmp(b.data, v.data(), v.size()) == 0);
  assert(b.rp_data_len == 2);
  assert(b.rp_data_lens[0] == v.size() && memcmp(b.rp_data[0], v.data(), v.size()) == 0);
  assert(b.rp_data_lens[1] == v2.size() && memcmp(b.rp_data[1], v2.data(), v2.size()) == 0);
}

int main() {
  test_empty();
  test_message();
  test_enumpb();
  test_nested();
  test_repeated();
  test_oneof();
  test_importing();
  test_bytes();

  std::cout << "C Test passed" << std::endl;
}