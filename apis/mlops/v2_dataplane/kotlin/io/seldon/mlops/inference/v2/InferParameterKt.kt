/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

//Generated by the protocol buffer compiler. DO NOT EDIT!
// source: v2_dataplane.proto

package io.seldon.mlops.inference.v2;

@kotlin.jvm.JvmSynthetic
public inline fun inferParameter(block: io.seldon.mlops.inference.v2.InferParameterKt.Dsl.() -> kotlin.Unit): io.seldon.mlops.inference.v2.V2Dataplane.InferParameter =
  io.seldon.mlops.inference.v2.InferParameterKt.Dsl._create(io.seldon.mlops.inference.v2.V2Dataplane.InferParameter.newBuilder()).apply { block() }._build()
public object InferParameterKt {
  @kotlin.OptIn(com.google.protobuf.kotlin.OnlyForUseByGeneratedProtoCode::class)
  @com.google.protobuf.kotlin.ProtoDslMarker
  public class Dsl private constructor(
    private val _builder: io.seldon.mlops.inference.v2.V2Dataplane.InferParameter.Builder
  ) {
    public companion object {
      @kotlin.jvm.JvmSynthetic
      @kotlin.PublishedApi
      internal fun _create(builder: io.seldon.mlops.inference.v2.V2Dataplane.InferParameter.Builder): Dsl = Dsl(builder)
    }

    @kotlin.jvm.JvmSynthetic
    @kotlin.PublishedApi
    internal fun _build(): io.seldon.mlops.inference.v2.V2Dataplane.InferParameter = _builder.build()

    /**
     * <pre>
     * A boolean parameter value.
     * </pre>
     *
     * <code>bool bool_param = 1;</code>
     */
    public var boolParam: kotlin.Boolean
      @JvmName("getBoolParam")
      get() = _builder.getBoolParam()
      @JvmName("setBoolParam")
      set(value) {
        _builder.setBoolParam(value)
      }
    /**
     * <pre>
     * A boolean parameter value.
     * </pre>
     *
     * <code>bool bool_param = 1;</code>
     */
    public fun clearBoolParam() {
      _builder.clearBoolParam()
    }
    /**
     * <pre>
     * A boolean parameter value.
     * </pre>
     *
     * <code>bool bool_param = 1;</code>
     * @return Whether the boolParam field is set.
     */
    public fun hasBoolParam(): kotlin.Boolean {
      return _builder.hasBoolParam()
    }

    /**
     * <pre>
     * An int64 parameter value.
     * </pre>
     *
     * <code>int64 int64_param = 2;</code>
     */
    public var int64Param: kotlin.Long
      @JvmName("getInt64Param")
      get() = _builder.getInt64Param()
      @JvmName("setInt64Param")
      set(value) {
        _builder.setInt64Param(value)
      }
    /**
     * <pre>
     * An int64 parameter value.
     * </pre>
     *
     * <code>int64 int64_param = 2;</code>
     */
    public fun clearInt64Param() {
      _builder.clearInt64Param()
    }
    /**
     * <pre>
     * An int64 parameter value.
     * </pre>
     *
     * <code>int64 int64_param = 2;</code>
     * @return Whether the int64Param field is set.
     */
    public fun hasInt64Param(): kotlin.Boolean {
      return _builder.hasInt64Param()
    }

    /**
     * <pre>
     * A string parameter value.
     * </pre>
     *
     * <code>string string_param = 3;</code>
     */
    public var stringParam: kotlin.String
      @JvmName("getStringParam")
      get() = _builder.getStringParam()
      @JvmName("setStringParam")
      set(value) {
        _builder.setStringParam(value)
      }
    /**
     * <pre>
     * A string parameter value.
     * </pre>
     *
     * <code>string string_param = 3;</code>
     */
    public fun clearStringParam() {
      _builder.clearStringParam()
    }
    /**
     * <pre>
     * A string parameter value.
     * </pre>
     *
     * <code>string string_param = 3;</code>
     * @return Whether the stringParam field is set.
     */
    public fun hasStringParam(): kotlin.Boolean {
      return _builder.hasStringParam()
    }
    public val parameterChoiceCase: io.seldon.mlops.inference.v2.V2Dataplane.InferParameter.ParameterChoiceCase
      @JvmName("getParameterChoiceCase")
      get() = _builder.getParameterChoiceCase()

    public fun clearParameterChoice() {
      _builder.clearParameterChoice()
    }
  }
}
@kotlin.jvm.JvmSynthetic
public inline fun io.seldon.mlops.inference.v2.V2Dataplane.InferParameter.copy(block: io.seldon.mlops.inference.v2.InferParameterKt.Dsl.() -> kotlin.Unit): io.seldon.mlops.inference.v2.V2Dataplane.InferParameter =
  io.seldon.mlops.inference.v2.InferParameterKt.Dsl._create(this.toBuilder()).apply { block() }._build()
