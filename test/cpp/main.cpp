#include <iostream>
#include <cassert>
#include <boost/json/src.hpp>

#include "empty.json.h"
#include "message.json.h"
#include "enumpb.json.h"
#include "nested.json.h"
#include "repeated.json.h"
#include "oneof.json.h"
#include "importing.json.h"
#include "bytes.json.h"
#include "jsonfield.json.h"
#include "optimistic.json.h"
#include "discard_if_default.json.h"
#include "no_serializer.json.h"

template<class T>
T identify(T v) {
  auto vs = jsonif::to_json(v);
  auto r = jsonif::from_json<T>(vs);
  auto rs = jsonif::to_json(v);
  assert(r == v);
  assert(rs == vs);
  return r;
}

void test_empty() {
  empty::Test a;
  a = identify(a);
}

void test_message() {
  message::Person a;
  assert(a.name == "");
  a = identify(a);
  assert(a.name == "");

  a.name = "foo";
  a = identify(a);
  assert(a.name == "foo");
}

void test_enumpb() {
  enumpb::Data a;
  assert(a == enumpb::Data::FOO);
  a = identify(a);
  assert(a == enumpb::Data::FOO);

  a = enumpb::Data::BAR;
  a = identify(a);
  assert(a == enumpb::Data::BAR);
}

void test_nested() {
  nested::nested::Test2 a;
  assert(a.nested_message.name == "");
  assert(a.nested_enum == nested::nested::Test::NestedEnum::FOO);
  assert(a.test.nested_message.name == "");
  assert(a.test.nested_enum == nested::nested::Test::NestedEnum::FOO);
  a = identify(a);
  assert(a.nested_message.name == "");
  assert(a.nested_enum == nested::nested::Test::NestedEnum::FOO);
  assert(a.test.nested_message.name == "");
  assert(a.test.nested_enum == nested::nested::Test::NestedEnum::FOO);

  a.nested_message.name = "foo";
  a.nested_enum = nested::nested::Test::NestedEnum::BAR;
  a.test.nested_message.name = "bar";
  a.test.nested_enum = nested::nested::Test::NestedEnum::HOGE;
  a = identify(a);
  assert(a.nested_message.name == "foo");
  assert(a.nested_enum == nested::nested::Test::NestedEnum::BAR);
  assert(a.test.nested_message.name == "bar");
  assert(a.test.nested_enum == nested::nested::Test::NestedEnum::HOGE);
}

void test_repeated() {
  repeated::Test a;
  assert(a.a.empty());
  assert(a.b.empty());
  assert(a.c.empty());
  assert(a.d.empty());
  a = identify(a);
  assert(a.a.empty());
  assert(a.b.empty());
  assert(a.c.empty());
  assert(a.d.empty());

  a.a.push_back(1);
  a.b.push_back("foo");
  a.c.push_back(repeated::Enum::BAR);
  a.d.push_back(repeated::Message{"bar"});
  a = identify(a);
  assert(a.a.size() == 1 && a.a.at(0) == 1);
  assert(a.b.size() == 1 && a.b.at(0) == "foo");
  assert(a.c.size() == 1 && a.c.at(0) == repeated::Enum::BAR);
  assert(a.d.size() == 1 && a.d.at(0).name == "bar");
}

void test_oneof() {
  oneof::Test a;
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::TEST_ONEOF_NOT_SET);
  a = identify(a);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::TEST_ONEOF_NOT_SET);

  a.set_a(1);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kA);
  assert(a.a == 1);
  a = identify(a);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kA);
  assert(a.a == 1);

  a.set_b("foo");
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kB);
  assert(a.b == "foo");
  a = identify(a);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kB);
  assert(a.b == "foo");

  a.set_c(oneof::Enum::BAR);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kC);
  assert(a.c == oneof::Enum::BAR);
  a = identify(a);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kC);
  assert(a.c == oneof::Enum::BAR);

  a.set_d(oneof::Message{"bar"});
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kD);
  assert(a.d.name == "bar");
  a = identify(a);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kD);
  assert(a.d.name == "bar");

  a.clear_test_oneof_case();
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::TEST_ONEOF_NOT_SET);
}

void test_importing() {
  importing::Test a;
  assert(a.t.nanos == 0);
  a = identify(a);
  assert(a.t.nanos == 0);
}

void test_bytes() {
  std::string v("\x00\x01\x02\x03", 4);
  std::string v2 = u8"???????????????";
  bytes::Test a;
  a.data = v;
  a.rp_data.push_back(v);
  a.rp_data.push_back(v2);
  a = identify(a);
  assert(a.data == v);
  assert(a.rp_data.at(0) == v);
  assert(a.rp_data.at(1) == v2);
}

void test_jsonfield() {
  jsonfield::Test a;
  a.field = 10;
  auto str = jsonif::to_json(a);
  assert(str == R"({"test":10,"hoge_field":0})");
  a = identify(a);
  assert(a.field == 10);
}

void test_optimistic() {
  auto str = R"({"b":"hoge"})";
  auto r = jsonif::from_json<optimistic::Test>(str);
  assert(r.a == "");
  assert(r.b == "hoge");
  r = identify(r);
  assert(r.a == "");
  assert(r.b == "hoge");
}

void test_discard_if_default() {
  discard_if_default::Test a;
  auto str = jsonif::to_json(a);
  assert(str == R"({"a":""})");

  a.b = "hoge";
  str = jsonif::to_json(a);
  assert(str == R"({"a":"","b":"hoge"})");

  a.b = "";
  a.c.a = 10;
  str = jsonif::to_json(a);
  assert(str == R"({"a":"","c":{"a":10}})");
}

namespace no_serializer {

static void tag_invoke(const boost::json::value_from_tag&, boost::json::value& jv, const ::no_serializer::Test& v) {
  boost::json::object obj;
  obj["b"] = boost::json::value_from(v.a);
  jv = std::move(obj);
}

static ::no_serializer::Test tag_invoke(const boost::json::value_to_tag<::no_serializer::Test>&, const boost::json::value& jv) {
  ::no_serializer::Test v;
  v.a = boost::json::value_to<std::string>(jv.at("b"));
  return v;
}

}

void test_no_serializer() {
  no_serializer::Test a;
  a.a = "hoge";
  auto str = jsonif::to_json(a);
  assert(str == R"({"b":"hoge"})");
  a = identify(a);
  assert(a.a == "hoge");
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
  test_jsonfield();
  test_optimistic();
  test_discard_if_default();
  test_no_serializer();

  std::cout << "C++ Test passed" << std::endl;
}